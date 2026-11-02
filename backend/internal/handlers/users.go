package handlers

import (
	"net/http"
	"strconv"
	"time"

	"admin_daebak_music/backend/internal/models"
	"admin_daebak_music/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UsersHandler struct {
	DB *gorm.DB
}

type createUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name"`
	Role     string `json:"role" binding:"required,oneof=admin customer"`
}

// Create lets an admin provision email/password accounts (register is disabled).
func (h *UsersHandler) Create(c *gin.Context) {
	var req createUserRequest
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
		Email:        req.Email,
		Username:     req.Username,
		Password:     hashed,
		FullName:     req.FullName,
		Role:         req.Role,
		AuthProvider: "email",
	}
	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email or username already exists"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user.ToResponse()})
}

func (h *UsersHandler) List(c *gin.Context) {
	var users []models.User
	if err := h.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}
	resp := make([]models.UserResponse, 0, len(users))
	for i := range users {
		resp = append(resp, users[i].ToResponse())
	}
	c.JSON(http.StatusOK, gin.H{"users": resp, "total": len(resp)})
}

func (h *UsersHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user.ToResponse()})
}

func (h *UsersHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("userID")
	var body struct {
		FullName  string `json:"full_name"`
		AvatarURL string `json:"avatar_url"`
		Username  string `json:"username"`
		Gender    string `json:"gender"`
		Birthdate string `json:"birthdate"` // YYYY-MM-DD
		Country   string `json:"country"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if body.FullName != "" {
		user.FullName = body.FullName
	}
	if body.AvatarURL != "" {
		user.AvatarURL = body.AvatarURL
	}
	if body.Username != "" {
		user.Username = body.Username
	}
	if body.Gender != "" {
		user.Gender = body.Gender
	}
	if body.Birthdate != "" {
		if t, err := time.Parse("2006-01-02", body.Birthdate); err == nil {
			user.Birthdate = &t
		}
	}
	if body.Country != "" {
		user.Country = body.Country
	}
	if err := h.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user.ToResponse()})
}