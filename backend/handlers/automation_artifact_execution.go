package handlers

import (
	"aiops-backend/database"
	"aiops-backend/models"
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func executeRepositoryArtifact(task models.AutomationTask, artifact repositoryArtifact, input []byte) (automationRunResult, error) {
	timeout := time.Duration(task.TimeoutSeconds) * time.Second
	switch task.NodeType {
	case "direct":
		return executeLocalRepositoryArtifact(task, artifact, input, timeout)
	case "ssh":
		return executeSSHRepositoryArtifact(task, artifact, input, timeout)
	case "agent":
		return executeAgentRepositoryArtifact(task, artifact, input, timeout)
	default:
		return automationRunResult{}, errors.New("不支持的节点类型")
	}
}

func executeLocalRepositoryArtifact(task models.AutomationTask, artifact repositoryArtifact, input []byte, timeout time.Duration) (automationRunResult, error) {
	root, err := os.MkdirTemp("", "aiops-task-artifact-")
	if err != nil {
		return automationRunResult{}, errors.New("创建任务临时目录失败")
	}
	defer os.RemoveAll(root)
	repositoryDir := filepath.Join(root, "repository")
	if err := os.MkdirAll(repositoryDir, 0700); err != nil {
		return automationRunResult{}, err
	}
	if err := extractRepositoryArtifact(artifact.Data, repositoryDir); err != nil {
		return automationRunResult{}, err
	}
	inputFile := filepath.Join(root, "input.json")
	if err := os.WriteFile(inputFile, input, 0600); err != nil {
		return automationRunResult{}, errors.New("写入任务输入失败")
	}
	workingDir := filepath.Join(repositoryDir, filepath.FromSlash(task.WorkingDirectory))
	if info, err := os.Stat(workingDir); err != nil || !info.IsDir() {
		return automationRunResult{}, errors.New("Git 工作目录不存在")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	command := exec.CommandContext(ctx, "bash", "-c", task.Command)
	command.Dir = workingDir
	command.Env = append(os.Environ(), "AIOPS_INPUT_FILE="+inputFile, "AIOPS_GIT_COMMIT="+artifact.Revision, "AIOPS_ARTIFACT_SHA256="+artifact.SHA256)
	var output synchronizedBuffer
	command.Stdout, command.Stderr = &output, &output
	err = command.Run()
	result := automationRunResult{Output: output.String(), ExitCode: processExitCode(err)}
	if ctx.Err() == context.DeadlineExceeded {
		result.ExitCode = -1
		return result, errors.New("任务执行超时")
	}
	return result, err
}

func executeSSHRepositoryArtifact(task models.AutomationTask, artifact repositoryArtifact, input []byte, timeout time.Duration) (automationRunResult, error) {
	var node models.CertificateExecutor
	if task.NodeID == 0 || database.DB.First(&node, task.NodeID).Error != nil {
		return automationRunResult{}, errors.New("SSH 节点不存在")
	}
	if node.Status != 1 {
		return automationRunResult{}, errors.New("SSH 节点已停用")
	}
	command := buildRemoteArtifactCommand(task, artifact, input)
	return runSSHCommandWithInput(node, command, artifact.Data, timeout)
}

func executeAgentRepositoryArtifact(task models.AutomationTask, artifact repositoryArtifact, input []byte, timeout time.Duration) (automationRunResult, error) {
	var node models.AgentConnection
	if task.NodeID == 0 || database.DB.First(&node, task.NodeID).Error != nil {
		return automationRunResult{}, errors.New("Agent 节点不存在")
	}
	if !agentVersionAtLeast(node.Version, minimumAgentArtifactVersion) {
		return automationRunResult{}, fmt.Errorf("Agent 版本过旧，请重新部署（Git 制品最低版本 %s）", minimumAgentArtifactVersion)
	}
	result, err := runAgentArtifact(node.ID, agentArtifactExecReq{ArtifactBase64: base64.StdEncoding.EncodeToString(artifact.Data), ArtifactSHA256: artifact.SHA256, Revision: artifact.Revision, WorkingDirectory: task.WorkingDirectory, Command: task.Command, InputJSON: string(input), TimeoutSeconds: task.TimeoutSeconds}, timeout)
	if err == nil && result.ExitCode != 0 {
		err = fmt.Errorf("命令退出码 %d", result.ExitCode)
	}
	output := result.Output
	if output == "" {
		output = strings.TrimSpace(result.Stdout + "\n" + result.Stderr)
	}
	return automationRunResult{Output: output, ExitCode: result.ExitCode}, err
}

func buildRemoteArtifactCommand(task models.AutomationTask, artifact repositoryArtifact, input []byte) string {
	workDir := "$AIOPS_TASK_DIR/repository"
	if task.WorkingDirectory != "" {
		workDir += "/" + task.WorkingDirectory
	}
	return strings.Join([]string{
		"set -e", "set -o pipefail", "AIOPS_TASK_DIR=\"$(mktemp -d /tmp/aiops-task-XXXXXX)\"", "trap 'rm -rf \"$AIOPS_TASK_DIR\"' EXIT",
		"cat > \"$AIOPS_TASK_DIR/artifact.tar.gz\"",
		"test \"$(sha256sum \"$AIOPS_TASK_DIR/artifact.tar.gz\" | awk '{print $1}')\" = " + shellQuote(artifact.SHA256),
		"mkdir -p \"$AIOPS_TASK_DIR/repository\"", "tar -xzf \"$AIOPS_TASK_DIR/artifact.tar.gz\" -C \"$AIOPS_TASK_DIR/repository\"",
		"printf '%s' " + shellQuote(base64.StdEncoding.EncodeToString(input)) + " | base64 -d > \"$AIOPS_TASK_DIR/input.json\"",
		"export AIOPS_INPUT_FILE=\"$AIOPS_TASK_DIR/input.json\"", "export AIOPS_GIT_COMMIT=" + shellQuote(artifact.Revision), "export AIOPS_ARTIFACT_SHA256=" + shellQuote(artifact.SHA256),
		"cd \"" + workDir + "\"", "bash -c " + shellQuote(task.Command),
	}, "\n")
}

func runSSHCommandWithInput(executor models.CertificateExecutor, command string, input []byte, timeout time.Duration) (automationRunResult, error) {
	secret, err := decryptSecret(executor.SecretCipher)
	if err != nil {
		return automationRunResult{}, errors.New("执行器认证信息解密失败")
	}
	auth, err := certificateSSHAuthMethod(executor.AuthType, secret)
	if err != nil {
		return automationRunResult{}, err
	}
	client, err := ssh.Dial("tcp", net.JoinHostPort(executor.Host, strconv.Itoa(defaultSSHPort(executor.Port))), &ssh.ClientConfig{User: executor.Username, Auth: []ssh.AuthMethod{auth}, HostKeyCallback: ssh.InsecureIgnoreHostKey(), Timeout: 10 * time.Second})
	if err != nil {
		return automationRunResult{}, err
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return automationRunResult{}, err
	}
	defer session.Close()
	var output synchronizedBuffer
	session.Stdin, session.Stdout, session.Stderr = bytes.NewReader(input), &output, &output
	done := make(chan error, 1)
	go func() { done <- session.Run(command) }()
	select {
	case err := <-done:
		exitCode := 0
		if err != nil {
			exitCode = 1
			var exitErr *ssh.ExitError
			if errors.As(err, &exitErr) {
				exitCode = exitErr.ExitStatus()
			}
		}
		return automationRunResult{Output: strings.TrimSpace(output.String()), ExitCode: exitCode}, err
	case <-time.After(timeout):
		_ = client.Close()
		return automationRunResult{Output: strings.TrimSpace(output.String()), ExitCode: -1}, errors.New("任务执行超时")
	}
}

func extractRepositoryArtifact(data []byte, destination string) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return errors.New("Git 制品格式不正确")
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errors.New("解压 Git 制品失败")
		}
		cleaned := filepath.Clean(filepath.FromSlash(header.Name))
		if cleaned == "." || cleaned == ".." || filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
			return errors.New("Git 制品包含非法路径")
		}
		target := filepath.Join(destination, cleaned)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode)&0777)
			if err != nil {
				return err
			}
			_, copyErr := io.CopyN(file, tarReader, header.Size)
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		case tar.TypeSymlink:
			link := filepath.Clean(filepath.FromSlash(header.Linkname))
			if filepath.IsAbs(link) || link == ".." || strings.HasPrefix(link, ".."+string(filepath.Separator)) {
				return errors.New("Git 制品包含非法符号链接")
			}
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.Symlink(link, target); err != nil {
				return err
			}
		default:
			return errors.New("Git 制品包含不支持的文件类型")
		}
	}
}
