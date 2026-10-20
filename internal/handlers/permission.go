package handlers

import (
	"guanlan/internal/database"
	"guanlan/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListPermissions(c *gin.Context) {
	var permissions []models.Permission
	database.DB.Order("id ASC").Find(&permissions)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "成功", "data": permissions})
}
