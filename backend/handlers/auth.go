package handlers

import (
	"aiops-backend/database"
	"aiops-backend/middleware"
	"aiops-backend/models"
	"aiops-backend/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	AuthType string `json:"auth_type"`
}

func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var user models.User
	var authErr error
	if strings.ToLower(strings.TrimSpace(req.AuthType)) == "ldap" {
		user, authErr = authenticateLDAP(req.Username, req.Password)
	} else {
		authErr = database.DB.Where("username = ?", req.Username).First(&user).Error
		if authErr == nil {
			authErr = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		}
	}
	if authErr != nil {
		middleware.RecordLogin(c, 0, req.Username, "失败", "用户名或密码错误")
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	if user.Status != 1 {
		middleware.RecordLogin(c, user.ID, user.Username, "失败", "账号已被禁用")
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "账号已被禁用"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成令牌失败"})
		return
	}
	middleware.RecordLogin(c, user.ID, user.Username, "成功", "登录成功")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "登录成功",
		"data": gin.H{
			"token": token,
		},
	})
}

func GetUserInfo(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := database.DB.Preload("Roles.Permissions").Preload("Roles.Menus").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	permSet := make(map[string]bool)
	menuSet := make(map[uint]models.Menu)

	for _, role := range user.Roles {
		if role.Status != 1 {
			continue
		}
		for _, perm := range role.Permissions {
			if perm.Status == 1 {
				permSet[perm.Code] = true
			}
		}
		for _, menu := range role.Menus {
			if menu.Status == 1 && menu.Type != 3 {
				menuSet[menu.ID] = menu
			}
		}
	}

	permissions := make([]string, 0, len(permSet))
	for code := range permSet {
		permissions = append(permissions, code)
	}

	menus := buildMenuTree(menuSet, 0)

	user.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"user":        user,
			"permissions": permissions,
			"menus":       menus,
		},
	})
}

func buildMenuTree(menuSet map[uint]models.Menu, parentID uint) []models.Menu {
	result := make([]models.Menu, 0)
	for _, menu := range menuSet {
		if menu.ParentID == parentID {
			menu.Children = buildMenuTree(menuSet, menu.ID)
			result = append(result, menu)
		}
	}
	// sort by sort field
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Sort > result[j].Sort {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func ChangePassword(c *gin.Context) {
	var req ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	userID := c.GetUint("user_id")
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "原密码错误"})
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	database.DB.Model(&user).Update("password", string(hashed))

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "密码修改成功"})
}
