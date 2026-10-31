package handlers

import (
	"net/http"
	"strconv"

	"admin_daebak_music/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UsersHandler struct {
	DB *gorm.DB
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
	if err := h.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user.ToResponse()})
}