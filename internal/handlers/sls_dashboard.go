package handlers

import (
	"fmt"
	"guanlan/internal/database"
	"guanlan/internal/models"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sls "github.com/alibabacloud-go/sls-20201230/v6/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gin-gonic/gin"
)

const (
	slsTicketEndpoint = "cn-shanghai.log.aliyuncs.com"
	slsTicketTTL      = int64(3600)
	maxSLSLogstores   = 10000
)

func ListSLSDashboardProjects(c *gin.Context) {
	var projects []models.SLSDashboardProject
	if err := database.DB.Order("sort ASC, id ASC").Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "读取 SLS 项目失败"})
		return
	}
	data := make([]slsDashboardProjectView, 0, len(projects))
	for _, project := range projects {
		provider := normalizeLogDashboardProvider(project.Provider)
		data = append(data, slsDashboardProjectView{
			ID: project.ID, Provider: provider, Name: project.Name, Project: project.Project, Region: project.Region, RoleArn: project.RoleArn,
			Description: project.Description, Sort: project.Sort,
			CredentialsConfigured: project.AccessKeyIDCipher != "" && project.AccessKeySecretCipher != "",
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": data})
}

func ListSLSDashboardLogstores(c *gin.Context) {
	project, auth, ok := loadSLSDashboardProject(c)
	if !ok {
		return
	}

	if normalizeLogDashboardProvider(project.Provider) == "tencent_cls" {
		topics, err := fetchTencentCLSTopics(auth, project.Region)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "获取 CLS Topic 失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": gin.H{
			"id": project.ID, "name": project.Name, "provider": "tencent_cls", "region": project.Region,
			"resource_type": "topic", "resources": topics,
		}})
		return
	}

	client, err := newSLSSDKClient(auth, slsRegionEndpoint(project.Region))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	logstores, err := fetchSLSLogstores(client, project.Project)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "获取 Logstore 失败: " + err.Error()})
		return
	}
	resources := make([]logDashboardResource, 0, len(logstores))
	for _, name := range logstores {
		resources = append(resources, logDashboardResource{ID: name, Name: name, Group: project.Project})
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": gin.H{
		"id": project.ID, "name": project.Name, "provider": "aliyun_sls", "project": project.Project, "region": project.Region,
		"resource_type": "logstore", "resources": resources, "logstores": logstores,
	}})
}

func CreateSLSDashboardSession(c *gin.Context) {
	project, auth, ok := loadSLSDashboardProject(c)
	if !ok {
		return
	}
	resourceID, err := url.PathUnescape(strings.TrimSpace(c.Param("logstore")))
	if err != nil || resourceID == "" || strings.Contains(resourceID, "/") {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "日志资源标识不正确"})
		return
	}
	if normalizeLogDashboardProvider(project.Provider) == "tencent_cls" {
		credentials, err := assumeTencentCLSRole(auth, project.Region, project.RoleArn)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "创建 CLS 临时会话失败: " + err.Error()})
			return
		}
		previewURL := buildTencentCLSRoleLoginURL(credentials, project.Region, resourceID, time.Now(), secureLoginNonce())
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": gin.H{
			"url": previewURL, "expires_in": tencentCLSRoleTTL, "topic_id": resourceID, "name": project.Name,
		}})
		return
	}

	client, err := newSLSSDKClient(auth, slsTicketEndpoint)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	runtime := newSLSRuntimeOptions()
	request := (&sls.CreateTicketRequest{}).
		SetExpirationTime(slsTicketTTL).
		SetAccessTokenExpirationTime(slsTicketTTL)
	response, err := client.CreateTicketWithOptions(request, map[string]*string{}, runtime)
	if err != nil || response == nil || response.Body == nil || strings.TrimSpace(tea.StringValue(response.Body.Ticket)) == "" {
		message := "阿里云未返回有效票据"
		if err != nil {
			message = err.Error()
		}
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "创建 SLS 免登录会话失败: " + message})
		return
	}

	previewURL := buildSLSConsoleURL(project.Project, resourceID, project.Region, tea.StringValue(response.Body.Ticket))
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": gin.H{
		"url":        previewURL,
		"expires_in": slsTicketTTL,
		"project":    project.Project,
		"logstore":   resourceID,
		"name":       project.Name,
	}})
}

func loadSLSDashboardProject(c *gin.Context) (models.SLSDashboardProject, slsAuth, bool) {
	id, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "SLS 项目 ID 不正确"})
		return models.SLSDashboardProject{}, slsAuth{}, false
	}
	var project models.SLSDashboardProject
	if err := database.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "SLS 项目不存在"})
		return models.SLSDashboardProject{}, slsAuth{}, false
	}
	auth, err := decryptSLSDashboardCredentials(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "SLS 凭证读取失败"})
		return models.SLSDashboardProject{}, slsAuth{}, false
	}
	return project, auth, true
}

func newSLSSDKClient(auth slsAuth, endpoint string) (*sls.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(strings.TrimSpace(auth.AccessKeyID)),
		AccessKeySecret: tea.String(strings.TrimSpace(auth.AccessKeySecret)),
		Endpoint:        tea.String(strings.TrimSpace(endpoint)),
		Protocol:        tea.String("HTTPS"),
	}
	if token := strings.TrimSpace(auth.SecurityToken); token != "" {
		config.SecurityToken = tea.String(token)
	}
	client, err := sls.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("初始化 SLS 客户端失败")
	}
	return client, nil
}

func newSLSRuntimeOptions() *util.RuntimeOptions {
	return (&util.RuntimeOptions{}).
		SetConnectTimeout(5000).
		SetReadTimeout(10000).
		SetAutoretry(true).
		SetMaxAttempts(3)
}

func fetchSLSLogstores(client *sls.Client, project string) ([]string, error) {
	const pageSize int32 = 500
	logstores := make([]string, 0)
	for offset := int32(0); len(logstores) < maxSLSLogstores; offset += pageSize {
		request := (&sls.ListLogStoresRequest{}).SetOffset(offset).SetSize(pageSize)
		response, err := client.ListLogStoresWithOptions(tea.String(strings.TrimSpace(project)), request, map[string]*string{}, newSLSRuntimeOptions())
		if err != nil {
			return nil, err
		}
		if response == nil || response.Body == nil {
			return nil, fmt.Errorf("阿里云未返回 Logstore 数据")
		}
		for _, item := range response.Body.Logstores {
			if name := strings.TrimSpace(tea.StringValue(item)); name != "" {
				logstores = append(logstores, name)
			}
		}
		if len(response.Body.Logstores) < int(pageSize) || len(logstores) >= int(tea.Int32Value(response.Body.Total)) {
			break
		}
	}
	sort.Strings(logstores)
	return logstores, nil
}

func slsRegionEndpoint(region string) string {
	return strings.TrimSpace(region) + ".log.aliyuncs.com"
}

func buildSLSConsoleURL(project, logstore, region, ticket string) string {
	query := url.Values{}
	query.Set("slsRegion", strings.TrimSpace(region))
	query.Set("sls_ticket", strings.TrimSpace(ticket))
	query.Set("isShare", "true")
	query.Set("hideTopbar", "true")
	query.Set("hideSidebar", "true")
	query.Set("ignoreTabLocalStorage", "true")
	return "https://sls.console.aliyun.com/lognext/project/" + url.PathEscape(strings.TrimSpace(project)) +
		"/logsearch/" + url.PathEscape(strings.TrimSpace(logstore)) + "?" + query.Encode()
}
