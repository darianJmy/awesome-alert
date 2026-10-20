package handlers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"guanlan/internal/database"
	"guanlan/internal/models"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type tunnelAuthConfig struct {
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	Passphrase string `json:"passphrase"`
}

type datasourceRequestTransport struct {
	base       http.RoundTripper
	authType   string
	authConfig string
	headers    map[string]string
}

func (t datasourceRequestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.Header = req.Header.Clone()
	for key, value := range t.headers {
		cloned.Header.Set(key, value)
	}
	setAuthHeader(cloned, t.authType, t.authConfig)
	return t.base.RoundTrip(cloned)
}

func datasourceRoundTripper(ds models.Datasource, base http.RoundTripper) http.RoundTripper {
	var config struct {
		CustomHeaders map[string]string `json:"custom_headers"`
	}
	_ = json.Unmarshal([]byte(ds.Config), &config)
	return datasourceRequestTransport{
		base:       base,
		authType:   ds.AuthType,
		authConfig: ds.AuthConfig,
		headers:    config.CustomHeaders,
	}
}

func datasourceHTTPClient(ds models.Datasource, timeout time.Duration) (*http.Client, func(), error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	cleanup := func() {
		transport.CloseIdleConnections()
	}

	if ds.TunnelID == 0 {
		return &http.Client{Timeout: timeout, Transport: datasourceRoundTripper(ds, transport)}, cleanup, nil
	}

	targetAddr, err := datasourceTargetAddr(ds.URL)
	if err != nil {
		return nil, cleanup, err
	}

	var tunnel models.Tunnel
	if err := database.DB.First(&tunnel, ds.TunnelID).Error; err != nil {
		return nil, cleanup, fmt.Errorf("隧道节点不存在: %w", err)
	}
	if tunnel.Status != 1 {
		return nil, cleanup, fmt.Errorf("隧道节点已禁用: %s", tunnel.Name)
	}
	if tunnel.Type != "" && tunnel.Type != "ssh" {
		return nil, cleanup, fmt.Errorf("隧道类型 %s 暂未接入执行器", tunnel.Type)
	}

	sshClient, err := openSSHClient(tunnel)
	if err != nil {
		return nil, cleanup, err
	}
	cleanup = func() {
		transport.CloseIdleConnections()
		_ = sshClient.Close()
	}

	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		type dialResult struct {
			conn net.Conn
			err  error
		}
		ch := make(chan dialResult, 1)
		go func() {
			conn, err := sshClient.Dial(network, targetAddr)
			ch <- dialResult{conn: conn, err: err}
		}()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case res := <-ch:
			return res.conn, res.err
		}
	}

	return &http.Client{Timeout: timeout, Transport: datasourceRoundTripper(ds, transport)}, cleanup, nil
}

func openSSHClient(t models.Tunnel) (*ssh.Client, error) {
	if strings.TrimSpace(t.Host) == "" {
		return nil, fmt.Errorf("隧道节点 Host 为空")
	}
	if t.Port <= 0 {
		t.Port = 22
	}
	if strings.TrimSpace(t.Username) == "" {
		return nil, fmt.Errorf("隧道节点用户名为空")
	}

	method, err := sshAuthMethod(t)
	if err != nil {
		return nil, err
	}
	cfg := &ssh.ClientConfig{
		User:            t.Username,
		Auth:            []ssh.AuthMethod{method},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         8 * time.Second,
	}
	addr := net.JoinHostPort(t.Host, strconv.Itoa(t.Port))
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("连接 SSH 隧道失败 %s: %w", addr, err)
	}
	return client, nil
}

func sshAuthMethod(t models.Tunnel) (ssh.AuthMethod, error) {
	var cfg tunnelAuthConfig
	_ = json.Unmarshal([]byte(t.AuthConfig), &cfg)

	switch t.AuthType {
	case "", "password":
		if cfg.Password == "" {
			return nil, fmt.Errorf("SSH 密码认证缺少 password")
		}
		return ssh.Password(cfg.Password), nil
	case "key":
		if cfg.PrivateKey == "" {
			return nil, fmt.Errorf("SSH Key 认证缺少 private_key")
		}
		var signer ssh.Signer
		var err error
		if cfg.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(cfg.PrivateKey), []byte(cfg.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
		}
		if err != nil {
			return nil, fmt.Errorf("解析 SSH 私钥失败: %w", err)
		}
		return ssh.PublicKeys(signer), nil
	default:
		return nil, fmt.Errorf("不支持的 SSH 认证方式: %s", t.AuthType)
	}
}

func datasourceTargetAddr(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", fmt.Errorf("数据源地址无效: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("HTTP 隧道要求数据源地址包含协议和 Host")
	}
	host := parsed.Hostname()
	port := parsed.Port()
	if host == "" {
		return "", fmt.Errorf("数据源地址 Host 为空")
	}
	if port == "" {
		switch parsed.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		default:
			return "", fmt.Errorf("HTTP 隧道暂不支持协议: %s", parsed.Scheme)
		}
	}
	return net.JoinHostPort(host, port), nil
}
