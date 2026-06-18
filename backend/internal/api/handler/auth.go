package handler

import (
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	var user model.StoreUser
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		middleware.Error(c, "用户名或密码错误")
		return
	}

	if user.Status != 1 {
		middleware.Error(c, "账号已被禁用")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		middleware.Error(c, "用户名或密码错误")
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, user.StoreID, user.Role)
	if err != nil {
		middleware.Error(c, "生成令牌失败")
		return
	}

	middleware.Success(c, dto.LoginResponse{
		Token: token,
		User: gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"real_name": user.RealName,
			"phone":     user.Phone,
			"store_id":  user.StoreID,
			"role":      user.Role,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	middleware.Success(c, nil)
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	username := middleware.GetUserID(c)
	storeID := middleware.GetStoreID(c)
	role := middleware.GetRole(c)

	middleware.Success(c, gin.H{
		"id":        userID,
		"username":  username,
		"store_id":  storeID,
		"role":      role,
	})
}
