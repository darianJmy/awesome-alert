package handlers

import (
	"fmt"
	"guanlan/internal/database"
	"guanlan/internal/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type saveAlertRuleGroupReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func ListAlertRuleGroups(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := database.DB.Model(&models.AlertRuleGroup{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询规则组失败"})
		return
	}
	var groups []models.AlertRuleGroup
	if err := query.Order("updated_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询规则组失败"})
		return
	}
	list := make([]gin.H, 0, len(groups))
	for _, group := range groups {
		list = append(list, alertRuleGroupResponse(group))
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": gin.H{
		"list": list, "total": total, "page": page, "page_size": pageSize,
	}})
}

func GetAlertRuleGroup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var group models.AlertRuleGroup
	if id < 1 || database.DB.First(&group, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "规则组不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": alertRuleGroupResponse(group)})
}

func CreateAlertRuleGroup(c *gin.Context) {
	var req saveAlertRuleGroupReq
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "规则组格式不正确"})
		return
	}
	name, description, err := normalizeAlertRuleGroupInput(req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if alertRuleGroupNameExists(name, 0) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "规则组名称已存在"})
		return
	}
	group := models.AlertRuleGroup{Name: name, Description: description}
	if database.DB.Create(&group).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建规则组失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "规则组已创建", "data": alertRuleGroupResponse(group)})
}

func UpdateAlertRuleGroup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var group models.AlertRuleGroup
	if id < 1 || database.DB.First(&group, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "规则组不存在"})
		return
	}
	var req saveAlertRuleGroupReq
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "规则组格式不正确"})
		return
	}
	name, description, err := normalizeAlertRuleGroupInput(req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if alertRuleGroupNameExists(name, group.ID) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "规则组名称已存在"})
		return
	}
	if err := database.DB.Model(&group).Updates(map[string]interface{}{"name": name, "description": description}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新规则组失败"})
		return
	}
	database.DB.First(&group, group.ID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "规则组已更新", "data": alertRuleGroupResponse(group)})
}

func DeleteAlertRuleGroup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var group models.AlertRuleGroup
	if id < 1 || database.DB.First(&group, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "规则组不存在"})
		return
	}
	var ruleCount int64
	database.DB.Model(&models.AlertRule{}).Where("group_id = ?", group.ID).Count(&ruleCount)
	if ruleCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请先删除规则组内的告警规则"})
		return
	}
	if database.DB.Delete(&group).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除规则组失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "规则组已删除"})
}

func normalizeAlertRuleGroupInput(rawName, rawDescription string) (string, string, error) {
	name, description := strings.TrimSpace(rawName), strings.TrimSpace(rawDescription)
	if name == "" || len([]rune(name)) > 128 {
		return "", "", fmt.Errorf("规则组名称不能为空且不能超过 128 个字符")
	}
	if len([]rune(description)) > 500 {
		return "", "", fmt.Errorf("规则组描述不能超过 500 个字符")
	}
	return name, description, nil
}

func alertRuleGroupNameExists(name string, excludeID uint) bool {
	query := database.DB.Model(&models.AlertRuleGroup{}).Where("name = ?", name)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	query.Count(&count)
	return count > 0
}

func validateDatasourceGroup(id uint) error {
	if id == 0 {
		return nil
	}
	var count int64
	database.DB.Model(&models.DatasourceGroup{}).Where("id = ?", id).Count(&count)
	if count == 0 {
		return fmt.Errorf("请选择有效的数据源分组")
	}
	var memberCount int64
	database.DB.Model(&models.DatasourceGroupMember{}).Where("group_id = ?", id).Count(&memberCount)
	if memberCount == 0 {
		return fmt.Errorf("数据源分组至少需要一个数据源")
	}
	return nil
}

func alertRuleGroupResponse(group models.AlertRuleGroup) gin.H {
	var ruleCount int64
	database.DB.Model(&models.AlertRule{}).Where("group_id = ?", group.ID).Count(&ruleCount)
	return gin.H{"id": group.ID, "name": group.Name, "description": group.Description, "rule_count": ruleCount, "created_at": group.CreatedAt, "updated_at": group.UpdatedAt}
}
