package handlers

import (
	"aiops-backend/database"
	"aiops-backend/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type saveDatasourceGroupReq struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	DatasourceIDs []uint `json:"datasource_ids"`
}

func ListDatasourceGroups(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	query := database.DB.Model(&models.DatasourceGroup{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", like, like)
	}
	var groups []models.DatasourceGroup
	if err := query.Order("created_at ASC, id ASC").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询数据源分组失败"})
		return
	}
	list := make([]gin.H, 0, len(groups))
	for _, group := range groups {
		list = append(list, datasourceGroupResponse(group))
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": list})
}

func CreateDatasourceGroup(c *gin.Context) {
	var req saveDatasourceGroupReq
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "数据源分组格式不正确"})
		return
	}
	name, description, datasourceIDs, err := normalizeDatasourceGroupInput(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if datasourceGroupNameExists(name, 0) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "数据源分组名称已存在"})
		return
	}
	group := models.DatasourceGroup{Name: name, Description: description}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&group).Error; err != nil {
			return err
		}
		return replaceDatasourceGroupMembers(tx, group.ID, datasourceIDs)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建数据源分组失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "数据源分组已创建", "data": datasourceGroupResponse(group)})
}

func UpdateDatasourceGroup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var group models.DatasourceGroup
	if id < 1 || database.DB.First(&group, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "数据源分组不存在"})
		return
	}
	var req saveDatasourceGroupReq
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "数据源分组格式不正确"})
		return
	}
	name, description, datasourceIDs, err := normalizeDatasourceGroupInput(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if datasourceGroupNameExists(name, group.ID) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "数据源分组名称已存在"})
		return
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&group).Updates(map[string]interface{}{"name": name, "description": description}).Error; err != nil {
			return err
		}
		if err := replaceDatasourceGroupMembers(tx, group.ID, datasourceIDs); err != nil {
			return err
		}
		return syncDatasourceGroupRuleBindings(tx, group.ID)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新数据源分组失败"})
		return
	}
	database.DB.First(&group, group.ID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "数据源分组已更新", "data": datasourceGroupResponse(group)})
}

func DeleteDatasourceGroup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var group models.DatasourceGroup
	if id < 1 || database.DB.First(&group, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "数据源分组不存在"})
		return
	}
	var usageCount int64
	database.DB.Model(&models.AlertRuleDatasourceGroup{}).Where("datasource_group_id = ?", group.ID).Count(&usageCount)
	if usageCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该数据源分组仍被告警规则使用，请先解除绑定"})
		return
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", group.ID).Delete(&models.DatasourceGroupMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&group).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除数据源分组失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "数据源分组已删除"})
}

func normalizeDatasourceGroupInput(req saveDatasourceGroupReq) (string, string, []uint, error) {
	name, description := strings.TrimSpace(req.Name), strings.TrimSpace(req.Description)
	if name == "" || len([]rune(name)) > 128 {
		return "", "", nil, fmt.Errorf("数据源分组名称不能为空且不能超过 128 个字符")
	}
	if len([]rune(description)) > 500 {
		return "", "", nil, fmt.Errorf("数据源分组描述不能超过 500 个字符")
	}
	datasourceIDs := normalizeChannelIDs(req.DatasourceIDs)
	if len(datasourceIDs) == 0 {
		return "", "", nil, fmt.Errorf("请至少选择一个 Prometheus 数据源")
	}
	var count int64
	database.DB.Model(&models.Datasource{}).Where("id IN ? AND type = ? AND status = 1", datasourceIDs, "prometheus").Count(&count)
	if count != int64(len(datasourceIDs)) {
		return "", "", nil, fmt.Errorf("数据源不存在、已停用或不是 Prometheus 类型")
	}
	return name, description, datasourceIDs, nil
}

func datasourceGroupNameExists(name string, excludeID uint) bool {
	query := database.DB.Model(&models.DatasourceGroup{}).Where("name = ?", name)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	query.Count(&count)
	return count > 0
}

func replaceDatasourceGroupMembers(tx *gorm.DB, groupID uint, datasourceIDs []uint) error {
	if err := tx.Where("group_id = ?", groupID).Delete(&models.DatasourceGroupMember{}).Error; err != nil {
		return err
	}
	for _, datasourceID := range datasourceIDs {
		if err := tx.Create(&models.DatasourceGroupMember{GroupID: groupID, DatasourceID: datasourceID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func datasourceGroupMemberIDs(tx *gorm.DB, groupID uint) ([]uint, error) {
	var members []models.DatasourceGroupMember
	if err := tx.Where("group_id = ?", groupID).Order("id ASC").Find(&members).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(members))
	for _, member := range members {
		ids = append(ids, member.DatasourceID)
	}
	return ids, nil
}

func syncDatasourceGroupRuleBindings(tx *gorm.DB, datasourceGroupID uint) error {
	var selections []models.AlertRuleDatasourceGroup
	if err := tx.Where("datasource_group_id = ?", datasourceGroupID).Find(&selections).Error; err != nil {
		return err
	}
	for _, selection := range selections {
		datasourceIDs, err := effectiveAlertRuleDatasourceIDs(tx, selection.RuleID)
		if err != nil {
			return err
		}
		if err := replaceAlertRuleBindings(tx, selection.RuleID, datasourceIDs); err != nil {
			return err
		}
	}
	return nil
}

func datasourceGroupResponse(group models.DatasourceGroup) gin.H {
	var members []models.DatasourceGroupMember
	database.DB.Where("group_id = ?", group.ID).Order("id ASC").Find(&members)
	datasourceIDs := make([]uint, 0, len(members))
	datasourceNames := make([]string, 0, len(members))
	for _, member := range members {
		var datasource models.Datasource
		database.DB.Select("id", "name").First(&datasource, member.DatasourceID)
		datasourceIDs = append(datasourceIDs, member.DatasourceID)
		datasourceNames = append(datasourceNames, datasource.Name)
	}
	var usageCount, ruleCount int64
	database.DB.Table("alert_rules").Joins("JOIN alert_rule_datasource_groups ON alert_rule_datasource_groups.rule_id = alert_rules.id").Where("alert_rule_datasource_groups.datasource_group_id = ?", group.ID).Distinct("alert_rules.group_id").Count(&usageCount)
	database.DB.Model(&models.AlertRuleDatasourceGroup{}).Where("datasource_group_id = ?", group.ID).Count(&ruleCount)
	return gin.H{"id": group.ID, "name": group.Name, "description": group.Description, "datasource_ids": datasourceIDs, "datasource_names": datasourceNames, "member_count": len(members), "usage_count": usageCount, "rule_count": ruleCount, "created_at": group.CreatedAt, "updated_at": group.UpdatedAt}
}
