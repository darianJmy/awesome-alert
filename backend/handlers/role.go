package handlers

import (
	"aiops-backend/database"
	"aiops-backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListRoles(c *gin.Context) {
	var roles []models.Role
	database.DB.Preload("Permissions").Preload("Menus").Order("id ASC").Find(&roles)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": roles})
}

type CreateRoleReq struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

func CreateRole(c *gin.Context) {
	var req CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if req.Status == 0 {
		req.Status = 1
	}

	role := models.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := database.DB.Create(&role).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "角色编码已存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "创建成功"})
}

type UpdateRoleReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

func UpdateRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	database.DB.Model(&models.Role{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"status":      req.Status,
	})

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
}

func DeleteRole(c *gin.Context) {
	id := c.Param("id")
	var role models.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "角色不存在"})
		return
	}
	database.DB.Select("Permissions", "Menus").Delete(&role)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

type AssignPermissionsReq struct {
	PermissionIDs []uint `json:"permission_ids"`
}

func AssignRolePermissions(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req AssignPermissionsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var role models.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "角色不存在"})
		return
	}

	var perms []models.Permission
	if len(req.PermissionIDs) > 0 {
		database.DB.Find(&perms, req.PermissionIDs)
	}
	database.DB.Model(&role).Association("Permissions").Replace(perms)

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "分配成功"})
}

type AssignMenusReq struct {
	MenuIDs []uint `json:"menu_ids"`
}

func AssignRoleMenus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req AssignMenusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var role models.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "角色不存在"})
		return
	}

	var menus []models.Menu
	if len(req.MenuIDs) > 0 {
		database.DB.Find(&menus, req.MenuIDs)
	}
	database.DB.Model(&role).Association("Menus").Replace(menus)

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "分配成功"})
}
