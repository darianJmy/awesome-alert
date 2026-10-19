package handlers

import (
	"errors"
	"fmt"
	"time"
)

func certificateExpiryAlertDays(value *int) (int, error) {
	if value == nil {
		return 30, nil
	}
	if *value < 1 || *value > 365 {
		return 0, errors.New("提前告警天数应为 1-365 天")
	}
	return *value, nil
}

func executeCertificateExpiryProbe(req ProbeReq) (networkProbeResult, int64, error) {
	days, err := certificateExpiryAlertDays(req.ExpiryAlertDays)
	if err != nil {
		return networkProbeResult{ProbeType: "certificate_expiry", CertificateStatus: "check_failed", Error: err.Error()}, 0, err
	}
	result, duration, err := executeSSLProbe(req)
	result.ProbeType = "certificate_expiry"
	result.ExpiryAlertDays = days
	if err != nil {
		result.CertificateStatus = "check_failed"
		result.Error = "无法获取证书：" + err.Error()
		return result, duration, errors.New(result.Error)
	}
	err = evaluateCertificateExpiry(&result, days, time.Now())
	return result, duration, err
}

// Compare timestamps, not truncated remaining days, at the alert boundary.
func evaluateCertificateExpiry(result *networkProbeResult, days int, now time.Time) error {
	if result.NotAfter.IsZero() {
		result.CertificateStatus = "check_failed"
		result.Error = "无法获取证书到期时间"
	} else {
		result.DaysRemaining = int(result.NotAfter.Sub(now).Hours() / 24)
		switch {
		case !result.NotAfter.After(now):
			result.CertificateStatus = "expired"
			result.Error = fmt.Sprintf("证书已过期，到期时间：%s", result.NotAfter.Format(time.RFC3339))
		case !result.NotAfter.After(now.Add(time.Duration(days) * 24 * time.Hour)):
			result.CertificateStatus = "expiring"
			result.Error = fmt.Sprintf("证书即将到期，剩余 %d 天（提前 %d 天告警），到期时间：%s", result.DaysRemaining, days, result.NotAfter.Format(time.RFC3339))
		default:
			result.CertificateStatus = "normal"
			result.Error = ""
			result.Output = fmt.Sprintf("证书正常，剩余 %d 天，到期时间：%s", result.DaysRemaining, result.NotAfter.Format(time.RFC3339))
			return nil
		}
	}
	result.Output = result.Error
	return errors.New(result.Error)
}
