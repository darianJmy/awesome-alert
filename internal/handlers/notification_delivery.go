package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	"guanlan/internal/config"
	"guanlan/internal/models"
)

type smtpNotificationConfig struct {
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	Security string   `json:"security"`
	Username string   `json:"username"`
	Password string   `json:"password,omitempty"`
	From     string   `json:"from"`
	To       []string `json:"to"`
}

func prepareNotificationChannel(req saveNotificationChannelReq, existing *models.NotificationChannel) (models.NotificationChannel, error) {
	var result models.NotificationChannel
	if notificationType(req.Type) == "email" {
		recipients := append([]string(nil), req.EmailRecipients...)
		if err := validateEmailRecipients(recipients); err != nil {
			return result, err
		}
		data, err := json.Marshal(recipients)
		if err != nil {
			return result, err
		}
		result.EmailRecipientsJSON = string(data)
		return result, nil
	}

	hook := strings.TrimSpace(req.WebhookURL)
	if hook == "" && existing != nil {
		result.WebhookURLCipher = existing.WebhookURLCipher
	} else {
		parsed, err := url.Parse(hook)
		if err != nil || parsed.Hostname() == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil {
			return result, errors.New("请填写有效的 HTTP/HTTPS Webhook 地址")
		}
		result.WebhookURLCipher, err = encryptWebhookURL(hook)
		if err != nil {
			return result, err
		}
	}
	if result.WebhookURLCipher == "" {
		return result, errors.New("请填写 Webhook 地址")
	}
	var err error
	result.SignSecretCipher, err = encryptSignSecret(req.Type, req.SignSecret, existing)
	return result, err
}

func validateSMTPConfig(c *smtpNotificationConfig) error {
	c.Host = strings.TrimSpace(c.Host)
	c.Username = strings.TrimSpace(c.Username)
	c.From = strings.TrimSpace(c.From)
	if c.Host == "" || strings.ContainsAny(c.Host, "/:@ \t\r\n") {
		return errors.New("请填写 SMTP 服务器主机名，不包含协议或端口")
	}
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("SMTP 端口必须为 1–65535")
	}
	switch c.Security {
	case "tls", "starttls", "none":
	default:
		return errors.New("请选择 SMTP 加密方式")
	}
	if strings.ContainsAny(c.Username, "\r\n") {
		return errors.New("SMTP 账号格式不正确")
	}
	if c.Username != "" && c.Password == "" {
		return errors.New("请填写 SMTP 密码或授权码")
	}
	if c.Username == "" {
		c.Password = ""
	}
	from, err := mail.ParseAddress(c.From)
	if err != nil || strings.ContainsAny(c.From, "\r\n") {
		return errors.New("发件人邮箱格式不正确")
	}
	c.From = from.String()
	return validateEmailRecipients(c.To)
}

func validateEmailRecipients(recipients []string) error {
	if len(recipients) == 0 {
		return errors.New("请至少填写一个收件人邮箱")
	}
	for i, value := range recipients {
		address, err := mail.ParseAddress(strings.TrimSpace(value))
		if err != nil || strings.ContainsAny(value, "\r\n") {
			return errors.New("收件人邮箱格式不正确")
		}
		recipients[i] = address.Address
	}
	return nil
}

func emailRecipients(channel models.NotificationChannel) []string {
	recipients := []string{}
	_ = json.Unmarshal([]byte(channel.EmailRecipientsJSON), &recipients)
	return recipients
}

func sharedSMTPConfig(channel models.NotificationChannel) (smtpNotificationConfig, error) {
	sender := config.AppConfig.SMTP
	c := smtpNotificationConfig{Host: sender.Host, Port: sender.Port, Security: sender.Security,
		Username: sender.Username, Password: sender.Password, From: sender.From, To: emailRecipients(channel)}
	if strings.TrimSpace(sender.Host) == "" {
		return c, errors.New("尚未配置全局邮件服务，请在配置文件的 smtp 部分填写发信信息")
	}
	if err := validateSMTPConfig(&c); err != nil {
		return c, fmt.Errorf("邮件配置不正确: %w", err)
	}
	return c, nil
}

func smtpMessage(config smtpNotificationConfig, title, content string) []byte {
	title = strings.ReplaceAll(strings.ReplaceAll(title, "\r", " "), "\n", " ")
	var msg bytes.Buffer
	fmt.Fprintf(&msg, "From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: base64\r\n\r\n", config.From, strings.Join(config.To, ", "), mime.BEncoding.Encode("UTF-8", title), time.Now().Format(time.RFC1123Z))
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	for len(encoded) > 76 {
		msg.WriteString(encoded[:76] + "\r\n")
		encoded = encoded[76:]
	}
	msg.WriteString(encoded + "\r\n")
	return msg.Bytes()
}

func sendSMTPNotification(channel models.NotificationChannel, body []byte) error {
	config, err := sharedSMTPConfig(channel)
	if err != nil {
		return err
	}
	if err := validateSMTPConfig(&config); err != nil {
		return err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}
	addr := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	dialer := &net.Dialer{Timeout: 8 * time.Second}
	tlsConfig := &tls.Config{ServerName: config.Host, MinVersion: tls.VersionTLS12}
	var conn net.Conn
	if config.Security == "tls" {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("SMTP 连接失败: %w", err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return fmt.Errorf("SMTP 握手失败: %w", err)
	}
	defer client.Close()
	if config.Security == "starttls" {
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("SMTP STARTTLS 失败: %w", err)
		}
	}
	if config.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", config.Username, config.Password, config.Host)); err != nil {
			return fmt.Errorf("SMTP 认证失败: %w", err)
		}
	}
	from, _ := mail.ParseAddress(config.From)
	if err := client.Mail(from.Address); err != nil {
		return fmt.Errorf("SMTP 发件人被拒绝: %w", err)
	}
	for _, to := range config.To {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("SMTP 收件人被拒绝: %w", err)
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = writer.Write(smtpMessage(config, notificationStringFromMap(data, "title"), notificationStringFromMap(data, "content"))); err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return fmt.Errorf("SMTP 邮件发送失败: %w", err)
	}
	// DATA was accepted; a later disconnect must not mark the delivered message as failed.
	_ = client.Quit()
	return nil
}

func sendDingTalkWebhook(channel models.NotificationChannel, body []byte) error {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}
	title := notificationStringFromMap(data, "title")
	if title == "" {
		title = "观澜运维平台通知"
	}
	payload, err := json.Marshal(map[string]interface{}{"msgtype": "markdown", "markdown": map[string]string{"title": title, "text": "### " + title + "\n\n" + notificationStringFromMap(data, "content")}})
	if err != nil {
		return err
	}
	if channel.SignSecretCipher != "" {
		hook, err := decryptSecret(channel.WebhookURLCipher)
		if err != nil {
			return err
		}
		secret, err := decryptSecret(channel.SignSecretCipher)
		if err != nil {
			return err
		}
		parsed, err := url.Parse(hook)
		if err != nil {
			return err
		}
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(timestamp + "\n" + secret))
		query := parsed.Query()
		query.Set("timestamp", timestamp)
		query.Set("sign", base64.StdEncoding.EncodeToString(mac.Sum(nil)))
		parsed.RawQuery = query.Encode()
		channel.WebhookURLCipher, err = encryptSecret(parsed.String())
		if err != nil {
			return err
		}
	}
	return postWebhook(channel, payload, "钉钉 Webhook")
}

func checkRobotResponse(kind string, body []byte) error {
	if kind != "feishu" && kind != "dingding" && kind != "qywechat" {
		return nil
	}
	var result struct {
		Code          *int   `json:"code"`
		ErrCode       *int   `json:"errcode"`
		StatusCode    *int   `json:"StatusCode"`
		Msg           string `json:"msg"`
		ErrMsg        string `json:"errmsg"`
		StatusMessage string `json:"StatusMessage"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return errors.New("机器人返回无效响应")
	}
	codes := []*int{result.ErrCode}
	message := result.ErrMsg
	if kind == "feishu" {
		codes = []*int{result.Code, result.StatusCode}
		message = result.Msg
		if message == "" {
			message = result.StatusMessage
		}
	}
	found := false
	for _, code := range codes {
		if code != nil {
			found = true
			if *code != 0 {
				return fmt.Errorf("机器人发送失败（%d）: %s", *code, message)
			}
		}
	}
	if !found {
		return errors.New("机器人响应缺少状态码")
	}
	return nil
}
