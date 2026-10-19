package handlers

import (
	"aiops-backend/database"
	"aiops-backend/models"
	"aiops-backend/utils"
	"archive/zip"
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createCertificateReq struct {
	Name        string   `json:"name" binding:"required"`
	Certificate string   `json:"certificate" binding:"required"`
	PrivateKey  string   `json:"private_key" binding:"required"`
	Domains     []string `json:"domains"`
	Description string   `json:"description"`
}

type updateCertificateReq struct {
	Certificate string `json:"certificate" binding:"required"`
	PrivateKey  string `json:"private_key" binding:"required"`
	Description string `json:"description"`
}

func ListCertificates(c *gin.Context) {
	var certs []models.Certificate
	tx := database.DB.Order("id DESC")
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("name LIKE ? OR domains LIKE ?", like, like)
	}
	tx.Find(&certs)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": certs})
}

func CreateCertificate(c *gin.Context) {
	var req createCertificateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请填写证书名称、公钥证书和私钥"})
		return
	}

	parsed, err := parseCertificateAndKey(req.Certificate, req.PrivateKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	domains := normalizeDomains(append(req.Domains, parsed.Domains...))
	domainsJSON, _ := json.Marshal(domains)
	keyCipher, err := encryptSecret(req.PrivateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "私钥加密失败"})
		return
	}

	versionName := time.Now().Format("20060102150405")
	now := time.Now()
	name := strings.TrimSpace(req.Name)
	var existing models.Certificate
	if err := database.DB.Where("name = ?", name).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "证书名称已存在，请换一个名称"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "校验证书名称失败"})
		return
	}

	cert := models.Certificate{
		Name:              name,
		Domains:           string(domainsJSON),
		Issuer:            parsed.Issuer,
		Subject:           parsed.Subject,
		Source:            "manual",
		SerialNumber:      parsed.SerialNumber,
		FingerprintSHA256: parsed.Fingerprint,
		NotBefore:         parsed.NotBefore,
		NotAfter:          parsed.NotAfter,
		Status:            certificateStatus(parsed.NotAfter),
		Description:       strings.TrimSpace(req.Description),
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&cert).Error; err != nil {
			return err
		}
		version := models.CertificateVersion{
			CertificateID:     cert.ID,
			Version:           versionName,
			CertPEM:           strings.TrimSpace(req.Certificate),
			PrivateKeyCipher:  keyCipher,
			Domains:           string(domainsJSON),
			Issuer:            parsed.Issuer,
			Subject:           parsed.Subject,
			SerialNumber:      parsed.SerialNumber,
			FingerprintSHA256: parsed.Fingerprint,
			NotBefore:         parsed.NotBefore,
			NotAfter:          parsed.NotAfter,
			Status:            "active",
			ActivatedAt:       &now,
		}
		if err := tx.Create(&version).Error; err != nil {
			return err
		}
		return tx.Model(&cert).Update("current_version_id", version.ID).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存证书失败: " + err.Error()})
		return
	}

	database.DB.First(&cert, cert.ID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "创建成功", "data": cert})
}

func ListCertificateVersions(c *gin.Context) {
	id := c.Param("id")
	var versions []models.CertificateVersion
	database.DB.Where("certificate_id = ?", id).Order("id DESC").Find(&versions)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": versions})
}

func UpdateCertificate(c *gin.Context) {
	var cert models.Certificate
	if err := database.DB.First(&cert, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "证书不存在"})
		return
	}

	var req updateCertificateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请填写公钥证书和私钥"})
		return
	}
	parsed, err := parseCertificateAndKey(req.Certificate, req.PrivateKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	keyCipher, err := encryptSecret(req.PrivateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "私钥加密失败"})
		return
	}

	domains := normalizeDomains(parsed.Domains)
	domainsJSON, _ := json.Marshal(domains)
	now := time.Now()
	versionName := now.Format("20060102150405")
	previousVersionID := cert.CurrentVersionID

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		version := models.CertificateVersion{
			CertificateID:     cert.ID,
			Version:           versionName,
			CertPEM:           strings.TrimSpace(req.Certificate),
			PrivateKeyCipher:  keyCipher,
			Domains:           string(domainsJSON),
			Issuer:            parsed.Issuer,
			Subject:           parsed.Subject,
			SerialNumber:      parsed.SerialNumber,
			FingerprintSHA256: parsed.Fingerprint,
			NotBefore:         parsed.NotBefore,
			NotAfter:          parsed.NotAfter,
			Status:            "active",
			ActivatedAt:       &now,
			PreviousVersionID: previousVersionID,
		}
		if err := tx.Create(&version).Error; err != nil {
			return err
		}
		if previousVersionID > 0 {
			tx.Model(&models.CertificateVersion{}).Where("id = ?", previousVersionID).Update("status", "inactive")
		}
		return tx.Model(&cert).Updates(map[string]interface{}{
			"domains":            string(domainsJSON),
			"issuer":             parsed.Issuer,
			"subject":            parsed.Subject,
			"serial_number":      parsed.SerialNumber,
			"fingerprint_sha256": parsed.Fingerprint,
			"not_before":         parsed.NotBefore,
			"not_after":          parsed.NotAfter,
			"status":             certificateStatus(parsed.NotAfter),
			"description":        strings.TrimSpace(req.Description),
			"current_version_id": version.ID,
		}).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新证书失败: " + err.Error()})
		return
	}

	database.DB.First(&cert, cert.ID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功", "data": cert})
}

func DownloadCertificate(c *gin.Context) {
	var cert models.Certificate
	if err := database.DB.First(&cert, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "证书不存在"})
		return
	}

	var version models.CertificateVersion
	tx := database.DB.Where("certificate_id = ?", cert.ID)
	if cert.CurrentVersionID > 0 {
		tx = tx.Where("id = ?", cert.CurrentVersionID)
	}
	if err := tx.Order("id DESC").First(&version).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "证书材料不存在"})
		return
	}

	privateKey, err := decryptSecret(version.PrivateKeyCipher)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "私钥解密失败"})
		return
	}

	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)
	if err := writeZipFile(zipWriter, "certificate.pem", strings.TrimSpace(version.CertPEM)+"\n"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成证书压缩包失败"})
		return
	}
	if err := writeZipFile(zipWriter, "private_key.pem", strings.TrimSpace(privateKey)+"\n"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成证书压缩包失败"})
		return
	}
	if err := zipWriter.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成证书压缩包失败"})
		return
	}

	filename := safeDownloadName(cert.Name) + ".zip"
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", filename, url.QueryEscape(filename)))
	c.Data(http.StatusOK, "application/zip", buffer.Bytes())
}

func DeleteCertificate(c *gin.Context) {
	id := c.Param("id")
	var taskNames []string
	database.DB.Model(&models.AutomationTask{}).Where("certificate_id = ?", id).Order("id ASC").Limit(6).Pluck("name", &taskNames)
	if len(taskNames) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "证书已被自动化任务引用，不能删除：" + strings.Join(taskNames, "、")})
		return
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("certificate_id = ?", id).Delete(&models.CertificateVersion{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Certificate{}, id).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除证书失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

func boundCertificateDomainNames(column string, id string) []string {
	var names []string
	database.DB.Model(&models.CertificateDomain{}).
		Where(column+" = ?", id).
		Order("id ASC").
		Limit(6).
		Pluck("domain", &names)
	return names
}

func boundResourceMessage(resource string, names []string) string {
	if len(names) == 0 {
		return resource + "已被域名绑定，不能删除"
	}
	visible := names
	suffix := ""
	if len(names) > 5 {
		visible = names[:5]
		suffix = " 等"
	}
	return fmt.Sprintf("%s已被域名绑定，不能删除：%s%s", resource, strings.Join(visible, "、"), suffix)
}

type saveDomainReq struct {
	Domain               string `json:"domain" binding:"required"`
	Port                 int    `json:"port"`
	FaultCenterID        uint   `json:"fault_center_id"`
	Owner                string `json:"owner"`
	Environment          string `json:"environment"`
	MonitorEnabled       *bool  `json:"monitor_enabled"`
	CheckIntervalMinutes int    `json:"check_interval_minutes"`
	ExpiryAlertDays      int    `json:"expiry_alert_days"`
	Severity             string `json:"severity"`
	Description          string `json:"description"`
}

func ListCertificateDomains(c *gin.Context) {
	var domains []models.CertificateDomain
	tx := database.DB.Model(&models.CertificateDomain{})
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("domain LIKE ? OR owner LIKE ?", like, like)
	}

	if _, ok := c.GetQuery("page"); !ok {
		tx.Order("id DESC").Find(&domains)
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": domains})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var total int64
	tx.Count(&total)
	tx.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&domains)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"list":      domains,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func CreateCertificateDomain(c *gin.Context) {
	var req saveDomainReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if err := validateCertificateDomainFaultCenter(req.FaultCenterID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	severity, err := validateCertificateDomainAlertPolicy(req.ExpiryAlertDays, req.Severity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	enabled := true
	if req.MonitorEnabled != nil {
		enabled = *req.MonitorEnabled
	}
	item := models.CertificateDomain{
		Domain:               normalizeTLSName(req.Domain),
		Port:                 defaultPort(req.Port),
		FaultCenterID:        req.FaultCenterID,
		Owner:                strings.TrimSpace(req.Owner),
		Environment:          strings.TrimSpace(req.Environment),
		MonitorEnabled:       enabled,
		CheckIntervalMinutes: defaultInterval(req.CheckIntervalMinutes),
		ExpiryAlertDays:      req.ExpiryAlertDays,
		Severity:             severity,
		AlertBeforeDays:      defaultAlertDays(""),
		Status:               "unknown",
		Description:          strings.TrimSpace(req.Description),
	}
	database.DB.Create(&item)
	if req.MonitorEnabled != nil && !*req.MonitorEnabled {
		database.DB.Model(&item).Update("monitor_enabled", false)
		item.MonitorEnabled = false
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "创建成功", "data": item})
}

func UpdateCertificateDomain(c *gin.Context) {
	id := c.Param("id")
	var req saveDomainReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if err := validateCertificateDomainFaultCenter(req.FaultCenterID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	severity, err := validateCertificateDomainAlertPolicy(req.ExpiryAlertDays, req.Severity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"domain":                  normalizeTLSName(req.Domain),
		"port":                    defaultPort(req.Port),
		"fault_center_id":         req.FaultCenterID,
		"certificate_id":          0,
		"executor_id":             0,
		"agent_id":                0,
		"deploy_script_id":        0,
		"notification_channel_id": 0,
		"deploy_params":           "",
		"deploy_cert_path":        "",
		"deploy_key_path":         "",
		"deploy_reload_command":   "",
		"owner":                   strings.TrimSpace(req.Owner),
		"environment":             strings.TrimSpace(req.Environment),
		"check_interval_minutes":  defaultInterval(req.CheckIntervalMinutes),
		"expiry_alert_days":       req.ExpiryAlertDays,
		"severity":                severity,
		"description":             strings.TrimSpace(req.Description),
	}
	if req.MonitorEnabled != nil {
		updates["monitor_enabled"] = *req.MonitorEnabled
	}
	database.DB.Model(&models.CertificateDomain{}).Where("id = ?", id).Updates(updates)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
}

func validateCertificateDomainFaultCenter(id uint) error {
	if id == 0 {
		return errors.New("请选择故障中心")
	}
	var count int64
	database.DB.Model(&models.FaultCenter{}).Where("id = ? AND status = ?", id, 1).Count(&count)
	if count == 0 {
		return errors.New("故障中心不存在或已停用")
	}
	return nil
}

func validateCertificateDomainAlertPolicy(expiryAlertDays int, severityValue string) (string, error) {
	if expiryAlertDays < 1 || expiryAlertDays > 365 {
		return "", errors.New("到期提醒应为 1-365 天")
	}
	severity := normalizeCertificateDomainSeverity(severityValue)
	if severity != "P0" && severity != "P1" && severity != "P2" {
		return "", errors.New("告警等级不正确")
	}
	return severity, nil
}

func DeleteCertificateDomain(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "域名 ID 无效"})
		return
	}
	if err := deleteCertificateDomain(database.DB, uint(id), time.Now()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除域名失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// Keep alert history while atomically ending monitoring for the deleted domain.
func deleteCertificateDomain(db *gorm.DB, id uint, now time.Time) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.CertificateDomain{}, id).Error; err != nil {
			return err
		}
		var alerts []models.AlertInstance
		if err := tx.Where("source = ? AND source_id = ? AND status IN ?", "certificate-domain", id,
			[]string{alertStatusPending, alertStatusFiring, alertStatusRecovering}).Find(&alerts).Error; err != nil {
			return err
		}
		for _, alert := range alerts {
			previous := alert.Status
			if err := tx.Model(&alert).Updates(map[string]interface{}{
				"status": alertStatusTerminated, "resolved_at": now, "state_changed_at": now, "last_event_at": now,
				"notification_status": notificationStatusTerminated, "notification_reason": "证书域名已删除", "next_notification_at": nil,
			}).Error; err != nil {
				return err
			}
			alert.Status, alert.ResolvedAt = alertStatusTerminated, &now
			if err := recordAlertEvent(tx, alert, "state_changed", previous, alertStatusTerminated,
				alert.Severity, alert.Severity, "certificate_domain_deleted", "", alert.LastProbeStatus, now); err != nil {
				return err
			}
		}
		return nil
	})
}

func CheckCertificateDomain(c *gin.Context) {
	var item models.CertificateDomain
	if err := database.DB.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "域名不存在"})
		return
	}
	if err := checkAndSaveDomain(&item); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error(), "data": item})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "采集完成", "data": item})
}

func CheckAllCertificateDomains(c *gin.Context) {
	var domains []models.CertificateDomain
	database.DB.Where("monitor_enabled = ?", true).Find(&domains)
	success := 0
	for i := range domains {
		if err := checkAndSaveDomain(&domains[i]); err == nil {
			success++
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": fmt.Sprintf("完成 %d/%d", success, len(domains)), "data": gin.H{"success": success, "total": len(domains)}})
}

func RunDueCertificateDomainChecks() {
	var domains []models.CertificateDomain
	database.DB.Where("monitor_enabled = ?", true).Find(&domains)
	now := time.Now()
	for i := range domains {
		if domains[i].LastCheckedAt != nil && now.Sub(*domains[i].LastCheckedAt) < time.Duration(defaultInterval(domains[i].CheckIntervalMinutes))*time.Minute {
			continue
		}
		_ = checkAndSaveDomain(&domains[i])
	}
}

type parsedCert struct {
	Domains      []string
	Issuer       string
	Subject      string
	SerialNumber string
	Fingerprint  string
	NotBefore    time.Time
	NotAfter     time.Time
}

func parseCertificateAndKey(certPEM, keyPEM string) (parsedCert, error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(certPEM)))
	if block == nil || block.Type != "CERTIFICATE" {
		return parsedCert{}, errors.New("公钥证书格式错误")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return parsedCert{}, errors.New("解析公钥证书失败，请确认上传的是 PEM 格式证书")
	}
	keyBlock, _ := pem.Decode([]byte(strings.TrimSpace(keyPEM)))
	if keyBlock == nil {
		return parsedCert{}, errors.New("私钥格式错误")
	}
	key, err := parsePrivateKey(keyBlock.Bytes)
	if err != nil {
		return parsedCert{}, err
	}
	if !publicKeyMatches(cert.PublicKey, key) {
		return parsedCert{}, errors.New("证书和私钥不匹配")
	}
	return parsedFromX509(cert), nil
}

func parsePrivateKey(der []byte) (interface{}, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return &key.PublicKey, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch k := key.(type) {
		case *rsa.PrivateKey:
			return &k.PublicKey, nil
		case *ecdsa.PrivateKey:
			return &k.PublicKey, nil
		case ed25519.PrivateKey:
			return k.Public(), nil
		default:
			return nil, errors.New("不支持的 PKCS8 私钥类型")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return &key.PublicKey, nil
	}
	return nil, errors.New("解析私钥失败，请确认上传的是 RSA、EC 或 Ed25519 PEM 私钥")
}

func publicKeyMatches(certKey, privatePublicKey interface{}) bool {
	a, errA := x509.MarshalPKIXPublicKey(certKey)
	b, errB := x509.MarshalPKIXPublicKey(privatePublicKey)
	return errA == nil && errB == nil && string(a) == string(b)
}

func parsedFromX509(cert *x509.Certificate) parsedCert {
	sum := sha256.Sum256(cert.Raw)
	domains := append([]string{}, cert.DNSNames...)
	for _, ip := range cert.IPAddresses {
		domains = append(domains, ip.String())
	}
	if len(domains) == 0 && cert.Subject.CommonName != "" {
		domains = append(domains, cert.Subject.CommonName)
	}
	return parsedCert{
		Domains:      normalizeDomains(domains),
		Issuer:       cert.Issuer.String(),
		Subject:      cert.Subject.String(),
		SerialNumber: cert.SerialNumber.String(),
		Fingerprint:  strings.ToUpper(hex.EncodeToString(sum[:])),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
	}
}

func checkAndSaveDomain(item *models.CertificateDomain) error {
	observed, err := fetchTLSCertificate(item.Domain, defaultPort(item.Port))
	now := time.Now()
	updates := map[string]interface{}{
		"last_checked_at": now,
	}
	if err != nil {
		updates["last_check_status"] = "failed"
		updates["last_check_error"] = err.Error()
		updates["status"] = "failed"
		database.DB.Model(item).Updates(updates)
		if err := database.DB.First(item, item.ID).Error; err != nil {
			return err
		}
		_ = syncCertificateDomainAlert(*item, now)
		return err
	}

	remaining := int(observed.NotAfter.Sub(now).Hours() / 24)
	status := certificateDomainStatus(observed.NotAfter, now, item.ExpiryAlertDays)
	updates["last_check_status"] = "success"
	updates["last_check_error"] = ""
	updates["last_seen_fingerprint"] = observed.Fingerprint
	updates["observed_issuer"] = observed.Issuer
	updates["observed_subject"] = observed.Subject
	updates["observed_not_before"] = observed.NotBefore
	updates["observed_expires_at"] = observed.NotAfter
	updates["days_remaining"] = remaining
	updates["last_observed_version_id"] = 0
	updates["status"] = status
	database.DB.Model(item).Updates(updates)
	if err := database.DB.First(item, item.ID).Error; err != nil {
		return err
	}
	_ = syncCertificateDomainAlert(*item, now)
	return nil
}

func fetchTLSCertificate(domain string, port int) (parsedCert, error) {
	name := normalizeTLSName(domain)
	address := net.JoinHostPort(name, strconv.Itoa(defaultPort(port)))
	dialer := &net.Dialer{Timeout: 8 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{ServerName: name, InsecureSkipVerify: true})
	if err != nil {
		return parsedCert{}, err
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return parsedCert{}, errors.New("未获取到线上证书")
	}
	return parsedFromX509(certs[0]), nil
}

func encryptSecret(plain string) (string, error) {
	return utils.EncryptSecret(plain)
}

func decryptSecret(ciphertext string) (string, error) {
	return utils.DecryptSecret(ciphertext)
}

func safeDownloadName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "certificate"
	}
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "-", "?", "-", "\"", "", "<", "-", ">", "-", "|", "-")
	return replacer.Replace(value)
}

func writeZipFile(writer *zip.Writer, name string, content string) error {
	file, err := writer.Create(name)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(content))
	return err
}

func certificateStatus(notAfter time.Time) string {
	days := int(time.Until(notAfter).Hours() / 24)
	if days < 0 {
		return "expired"
	}
	if days <= 30 {
		return "expiring"
	}
	return "normal"
}

func certificateDomainStatus(notAfter, now time.Time, expiryAlertDays int) string {
	if notAfter.Before(now) {
		return "expired"
	}
	if !notAfter.After(now.AddDate(0, 0, expiryAlertDays)) {
		return "expiring"
	}
	return "normal"
}

func normalizeCertificateDomainSeverity(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "P2"
	}
	return value
}

func normalizeTLSName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "http://")
	value = strings.Split(value, "/")[0]
	host, _, err := net.SplitHostPort(value)
	if err == nil {
		return host
	}
	return value
}

func defaultPort(port int) int {
	if port <= 0 {
		return 443
	}
	return port
}

func defaultInterval(minutes int) int {
	if minutes <= 0 {
		return 360
	}
	return minutes
}

func defaultAlertDays(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "30,15,7,3,1"
	}
	return value
}
