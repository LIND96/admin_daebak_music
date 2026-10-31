package handlers

import (
	"net/http"
	"time"

	"admin_daebak_music/backend/internal/config"
	"admin_daebak_music/backend/internal/models"
	"admin_daebak_music/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Auth cookie carries the same JWT previously returned in JSON and kept in
// localStorage. HttpOnly keeps it out of JS reach (XSS), SameSite=Lax blocks
// cross-site POST from sending it (CSRF), Secure (prod) forces HTTPS.
const authCookieName = "token"

func cookieMaxAge(expiry string) int {
	dur, err := time.ParseDuration(expiry)
	if err != nil {
		dur = 24 * time.Hour
	}
	return int(dur.Seconds())
}

func setAuthCookie(c *gin.Context, cfg *config.Config, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(authCookieName, token, cookieMaxAge(cfg.JWTExpiry), "/", "", cfg.CookieSecure, true)
}

func clearAuthCookie(c *gin.Context, cfg *config.Config) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(authCookieName, "", -1, "/", "", cfg.CookieSecure, true)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: hashed,
		FullName: req.FullName,
		Role:     "user",
	}
	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email or username already exists"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, h.Cfg.JWTSecret, h.Cfg.JWTExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	setAuthCookie(c, h.Cfg, token)
	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  user.ToResponse(),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !utils.CheckPassword(user.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, h.Cfg.JWTSecret, h.Cfg.JWTExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	setAuthCookie(c, h.Cfg, token)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user.ToResponse(),
	})
}

// Logout is stateless with JWT; client discards token. Endpoint kept for API completeness.
func (h *AuthHandler) Logout(c *gin.Context) {
	clearAuthCookie(c, h.Cfg)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetUint("userID")
	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user.ToResponse()})
}