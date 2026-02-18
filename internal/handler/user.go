package handler

import (
	"Sparrow/internal/model"
	"Sparrow/internal/service"
	"Sparrow/internal/utils"
	"Sparrow/internal/utils/uJwt"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")

	post, err := h.service.GetUser(idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, post)
}

func (h *UserHandler) UserRegister(c *gin.Context) {
	var req struct {
		Nickname string `json:"nickname"`
		RealName string `json:"realName"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Birthday int64  `json:"birthday"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Log.Error("bind json err", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	BirthdayTime := time.Unix(req.Birthday, 0)

	post, err := h.service.UserRegister(req.Nickname, req.RealName, req.Email, req.Password, &BirthdayTime)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyRegistered) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		utils.Log.Error("user register err", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func (h *UserHandler) LogIn(c *gin.Context) {
	var userLogin model.UserLogin
	if err := c.BindJSON(&userLogin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.GetUserByEmail(userLogin.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		utils.Log.Error("get user by email err", zap.String("handle", "login"), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}
	ok, needsUpgrade := utils.VerifyPassword(user.Password, userLogin.Password)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	if needsUpgrade {
		if err := h.service.UpgradePasswordHash(user.ID, userLogin.Password); err != nil {
			utils.Log.Warn("failed to upgrade legacy password hash", zap.Int("userID", user.ID), zap.Error(err))
		}
	}

	token, err := uJwt.GenerateJWT(strconv.FormatInt(int64(user.ID), 10), user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
