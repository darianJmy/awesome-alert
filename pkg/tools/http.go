package tools

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"time"
)

// newClient 创建一个统一的 http.Client，默认忽略证书校验
func newClient(timeout int) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
	}
	return &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}
}

// doRequest 通用请求函数
func doRequest(method, url string, body *bytes.Reader, headers map[string]string, timeout int) (*http.Response, error) {
	client := newClient(timeout)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// 设置 header
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Get 请求
func Get(url string, headers map[string]string, timeout int) (*http.Response, error) {
	return doRequest(http.MethodGet, url, nil, headers, timeout)
}

// Post 请求
func Post(url string, body *bytes.Reader, headers map[string]string, timeout int) (*http.Response, error) {
	// 默认 JSON 类型（外部可覆盖）
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}
	return doRequest(http.MethodPost, url, body, headers, timeout)
}

// CreateBasicAuthHeader 创建带认证的HTTP头
func CreateBasicAuthHeader(username, password string) map[string]string {
	if username == "" && password == "" {
		return nil
	}
	return map[string]string{
		"Authorization": "Basic " + basicAuth(username, password),
	}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
