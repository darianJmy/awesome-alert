package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"guanlan/internal/models"
)

func TestCertificateExpiryBoundaries(t *testing.T) {
	now := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name   string
		expiry time.Time
		status string
	}{
		{"normal", now.Add(30*24*time.Hour + time.Second), "normal"},
		{"threshold", now.Add(30 * 24 * time.Hour), "expiring"},
		{"less than a day", now.Add(time.Hour), "expiring"},
		{"expires now", now, "expired"},
		{"expired", now.Add(-time.Hour), "expired"},
		{"missing", time.Time{}, "check_failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := networkProbeResult{NotAfter: tc.expiry}
			err := evaluateCertificateExpiry(&result, 30, now)
			if result.CertificateStatus != tc.status || (err == nil) != (tc.status == "normal") {
				t.Fatalf("unexpected result: %+v, %v", result, err)
			}
		})
	}
	if days, err := certificateExpiryAlertDays(nil); err != nil || days != 30 {
		t.Fatalf("default: %d %v", days, err)
	}
	for _, days := range []int{-1, 0, 366} {
		if _, err := certificateExpiryAlertDays(&days); err == nil {
			t.Fatalf("accepted invalid threshold %d", days)
		}
	}
}

func TestCertificateExpiryProbeAndTask(t *testing.T) {
	for _, tc := range []struct {
		name      string
		remaining time.Duration
		status    string
	}{
		{"normal", 60 * 24 * time.Hour, "normal"}, {"expiring", 10 * 24 * time.Hour, "expiring"}, {"expired", -24 * time.Hour, "expired"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			template := x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "localhost"}, DNSNames: []string{"localhost"}, NotBefore: time.Now().Add(-90 * 24 * time.Hour), NotAfter: time.Now().Add(tc.remaining), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
			der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
			if err != nil {
				t.Fatal(err)
			}
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
			server.TLS = &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}}
			server.StartTLS()
			defer server.Close()
			host, rawPort, _ := net.SplitHostPort(strings.TrimPrefix(server.URL, "https://"))
			port, _ := strconv.Atoi(rawPort)
			days := 30
			req := ProbeReq{Type: "certificate_expiry", Host: host, Port: port, ServerName: "localhost", ExpiryAlertDays: &days, TimeoutMS: 1000}
			config, _ := json.Marshal(req)
			status, raw, _ := executeInspection(models.InspectionItem{Type: "certificate_expiry", Config: string(config)})
			var result networkProbeResult
			if err := json.Unmarshal([]byte(raw), &result); err != nil {
				t.Fatal(err)
			}
			if result.CertificateStatus != tc.status || result.NotAfter.IsZero() || result.ExpiryAlertDays != 30 {
				t.Fatalf("certificate data lost: %s", raw)
			}
			if (status == "success") != (tc.status == "normal") {
				t.Fatalf("wrong task status %s", status)
			}
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/instant-probe", strings.NewReader(string(config)))
			ctx.Request.Header.Set("Content-Type", "application/json")
			InstantProbe(ctx)
			var response struct {
				Data struct {
					Status            string `json:"status"`
					CertificateStatus string `json:"certificate_status"`
					Result            string `json:"result"`
				} `json:"data"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Data.Status != status || response.Data.CertificateStatus != tc.status || !strings.Contains(response.Data.Result, "not_after") {
				t.Fatalf("instant probe differs: %s", recorder.Body.String())
			}
			// The existing SSL probe still succeeds with expired or self-signed certificates.
			if _, _, err := executeSSLProbe(req); err != nil {
				t.Fatalf("SSL behavior changed: %v", err)
			}
		})
	}
}

func TestCertificateExpiryConnectionFailureIsNotExpired(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	result, _, err := executeCertificateExpiryProbe(ProbeReq{Host: "127.0.0.1", Port: port, TimeoutMS: 100})
	if err == nil || result.CertificateStatus != "check_failed" || !result.NotAfter.IsZero() {
		t.Fatalf("connection failure misclassified: %+v %v", result, err)
	}
}

func TestCertificateExpiryConfigRoundTrip(t *testing.T) {
	item := models.InspectionItem{Name: "证书到期", Type: "certificate_expiry", Target: "example.com", Config: `{"host":"example.com","port":443,"expiry_alert_days":14,"execution_nodes":[{"node_type":"direct"}]}`, IntervalSeconds: 43200, FailureThreshold: 2, RecoveryThreshold: 1, FaultCenterID: 1, Severity: "P2"}
	doc, err := inspectionTaskConfigDocumentFromItems([]models.InspectionItem{item})
	if err != nil {
		t.Fatal(err)
	}
	content, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := parseInspectionTaskConfigDocument(string(content))
	if err != nil {
		t.Fatal(err)
	}
	task := restored.Tasks[0]
	raw, err := json.Marshal(task.Config)
	if err != nil {
		t.Fatal(err)
	}
	var config ProbeReq
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatal(err)
	}
	if config.ExpiryAlertDays == nil || *config.ExpiryAlertDays != 14 || task.Type != "certificate_expiry" || task.IntervalSeconds != 43200 || task.FailureThreshold != 2 {
		t.Fatalf("policy lost in export/import: %+v", task)
	}
}
