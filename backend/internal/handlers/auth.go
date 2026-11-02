package handlers

import (
	"net/http"
	"strconv"
	"strings"
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

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type bootstrapAdminRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required,min=6"`
	FullName  string `json:"full_name"`
	SetupKey  string `json:"setup_key" binding:"required"`
}

type googleLoginRequest struct {
	IDToken    string `json:"id_token" binding:"required"`
	DeviceInfo string `json:"device_info"`
	// Profil tambahan yang dikumpulkan HP saat login.
	Username    string `json:"username"`
	PhotoURL    string `json:"photo_url"`
	Gender      string `json:"gender"`
	Birthdate   string `json:"birthdate"` // YYYY-MM-DD
	Country     string `json:"country"`
	DeviceBrand string `json:"device_brand"`
	DeviceModel string `json:"device_model"`
	OSVersion   string `json:"os_version"`
	Platform    string `json:"platform"` // android | ios
}

// applyCustomerProfile fills mutable profile fields from a Google login request.
// Identity fields (username/photo/gender/birthdate) are set when provided and
// not already set; device/country fields are refreshed on every login.
func applyCustomerProfile(user *models.User, req googleLoginRequest) {
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.PhotoURL != "" {
		user.AvatarURL = req.PhotoURL
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	if req.Birthdate != "" {
		if t, err := time.Parse("2006-01-02", req.Birthdate); err == nil {
			user.Birthdate = &t
		}
	}
	if req.Country != "" {
		user.Country = req.Country
	}
	if req.DeviceBrand != "" {
		user.DeviceBrand = req.DeviceBrand
	}
	if req.DeviceModel != "" {
		user.DeviceModel = req.DeviceModel
	}
	if req.OSVersion != "" {
		user.OSVersion = req.OSVersion
	}
	if req.Platform != "" {
		user.Platform = req.Platform
	}
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Auth cookies: access JWT (short) + opaque refresh token (long, role-based).
// HttpOnly keeps them out of JS reach (XSS), SameSite=Lax blocks
// cross-site POST from sending them (CSRF), Secure (prod) forces HTTPS.
const authCookieName = "token"
const refreshCookieName = "refresh_token"

func parseExpiry(expiry string, fallback time.Duration) time.Duration {
	dur, err := time.ParseDuration(expiry)
	if err != nil {
		return fallback
	}
	return dur
}

func cookieMaxAge(dur time.Duration) int {
	return int(dur.Seconds())
}

// refreshExpiryFor returns the refresh-token lifetime by role.
// Customer (mobile) gets 90 days; admin stays short per policy.
func (h *AuthHandler) refreshExpiryFor(role string) time.Duration {
	if role == "admin" {
		return parseExpiry(h.Cfg.RefreshExpiryAdmin, 24*time.Hour)
	}
	return parseExpiry(h.Cfg.RefreshExpiryCustomer, 90*24*time.Hour)
}

func setAuthCookies(c *gin.Context, cfg *config.Config, access, refresh string, refreshDur time.Duration) {
	c.SetSameSite(http.SameSiteLaxMode)
	accessDur := parseExpiry(cfg.JWTExpiry, time.Hour)
	c.SetCookie(authCookieName, access, cookieMaxAge(accessDur), "/", "", cfg.CookieSecure, true)
	c.SetCookie(refreshCookieName, refresh, cookieMaxAge(refreshDur), "/api/auth", "", cfg.CookieSecure, true)
}

func clearAuthCookies(c *gin.Context, cfg *config.Config) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(authCookieName, "", -1, "/", "", cfg.CookieSecure, true)
	c.SetCookie(refreshCookieName, "", -1, "/api/auth", "", cfg.CookieSecure, true)
}

// issueSession creates an access JWT + opaque refresh token and persists the hash.
func (h *AuthHandler) issueSession(c *gin.Context, user *models.User) (access, refreshRaw string, refreshDur time.Duration, err error) {
	access, err = utils.GenerateToken(user.ID, user.Username, user.Role, h.Cfg.JWTSecret, h.Cfg.JWTExpiry)
	if err != nil {
		return "", "", 0, err
	}
	refreshDur = h.refreshExpiryFor(user.Role)
	refreshRaw, err = utils.GenerateRefreshToken()
	if err != nil {
		return "", "", 0, err
	}
	rt := models.RefreshToken{
		UserID:    user.ID,
		TokenHash: utils.HashRefreshToken(refreshRaw),
		ExpiresAt: time.Now().Add(refreshDur),
		DeviceInfo: c.GetHeader("User-Agent"),
	}
	if err := h.DB.Create(&rt).Error; err != nil {
		return "", "", 0, err
	}
	return access, refreshRaw, refreshDur, nil
}

func sessionResponse(access, refresh string, refreshDur time.Duration, user models.User) gin.H {
	return gin.H{
		"access_token":  access,
		"refresh_token": refresh,
		"token":         access, // backward compat for existing web client
		"expires_in":    int(parseExpiry("1h", time.Hour).Seconds()),
		"refresh_expires_in": int(refreshDur.Seconds()),
		"user":          user.ToResponse(),
	}
}

// BootstrapAdmin creates the first admin via a one-time setup key.
// Disabled (403) once any admin exists or when BOOTSTRAP_SECRET is unset.
func (h *AuthHandler) BootstrapAdmin(c *gin.Context) {
	var req bootstrapAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.Cfg.BootstrapSecret == "" || req.SetupKey != h.Cfg.BootstrapSecret {
		c.JSON(http.StatusForbidden, gin.H{"error": "bootstrap disabled"})
		return
	}
	var adminCount int64
	if err := h.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check admins"})
		return
	}
	if adminCount > 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin already exists"})
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
		Role:         "admin",
		AuthProvider: "email",
	}
	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email or username already exists"})
		return
	}

	access, refreshRaw, refreshDur, err := h.issueSession(c, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate session"})
		return
	}

	setAuthCookies(c, h.Cfg, access, refreshRaw, refreshDur)
	c.JSON(http.StatusCreated, sessionResponse(access, refreshRaw, refreshDur, user))
}

// Login is admin-only (web). Customers must use Google login.
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

	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only: use google login for customer accounts"})
		return
	}

	if user.Password == "" || !utils.CheckPassword(user.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	access, refreshRaw, refreshDur, err := h.issueSession(c, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate session"})
		return
	}

	setAuthCookies(c, h.Cfg, access, refreshRaw, refreshDur)
	c.JSON(http.StatusOK, sessionResponse(access, refreshRaw, refreshDur, user))
}

// GoogleLogin verifies a mobile ID token and issues our own session.
// Admins are rejected here; they must use email+password login.
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req googleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := utils.VerifyGoogleIDToken(c.Request.Context(), req.IDToken,
		h.Cfg.GoogleAndroidClientID, h.Cfg.GoogleIOSClientID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid google id_token: " + err.Error()})
		return
	}

	var user models.User
	// 1) match by google_sub, 2) link by email, 3) create customer.
	if err := h.DB.Where("google_sub = ?", profile.Sub).First(&user).Error; err != nil {
		if err := h.DB.Where("email = ?", profile.Email).First(&user).Error; err != nil {
			username := strings.ToLower(strings.Split(profile.Email, "@")[0])
			if req.Username != "" {
				username = req.Username
			}
			base := username
			for i := 1; ; i++ {
				var n int64
				h.DB.Model(&models.User{}).Where("username = ?", username).Count(&n)
				if n == 0 {
					break
				}
				username = base + "-" + strconv.Itoa(i)
			}
			sub := profile.Sub
			user = models.User{
				Email:        profile.Email,
				Username:     username,
				FullName:     profile.Name,
				AvatarURL:    profile.Picture,
				Role:         "customer",
				AuthProvider: "google",
				GoogleSub:    &sub,
			}
			applyCustomerProfile(&user, req)
			if err := h.DB.Create(&user).Error; err != nil {
				c.JSON(http.StatusConflict, gin.H{"error": "email or username already exists"})
				return
			}
		} else {
			// Link existing email account to this Google subject.
			if user.Role == "admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "admin must login with email and password"})
				return
			}
			sub := profile.Sub
			user.GoogleSub = &sub
			if user.AuthProvider == "" {
				user.AuthProvider = "google"
			}
			if user.AvatarURL == "" {
				user.AvatarURL = profile.Picture
			}
			if user.FullName == "" {
				user.FullName = profile.Name
			}
			applyCustomerProfile(&user, req)
			if err := h.DB.Save(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link google account"})
				return
			}
		}
	} else {
		// Returning customer: refresh collected device/country data.
		applyCustomerProfile(&user, req)
		if err := h.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
			return
		}
	}

	if user.Role == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin must login with email and password"})
		return
	}

	access, refreshRaw, refreshDur, err := h.issueSession(c, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate session"})
		return
	}

	setAuthCookies(c, h.Cfg, access, refreshRaw, refreshDur)
	c.JSON(http.StatusOK, sessionResponse(access, refreshRaw, refreshDur, user))
}

// Refresh rotates the refresh token and issues a new pair.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	_ = c.ShouldBindJSON(&req)
	raw := req.RefreshToken
	if raw == "" {
		if ck, err := c.Cookie(refreshCookieName); err == nil {
			raw = ck
		}
	}
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing refresh_token"})
		return
	}

	var stored models.RefreshToken
	if err := h.DB.Where("token_hash = ?", utils.HashRefreshToken(raw)).First(&stored).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	if stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired or revoked"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, stored.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Rotate: revoke old, issue new.
	now := time.Now()
	stored.RevokedAt = &now
	if err := h.DB.Save(&stored).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rotate token"})
		return
	}

	access, refreshRaw, refreshDur, err := h.issueSession(c, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate session"})
		return
	}

	setAuthCookies(c, h.Cfg, access, refreshRaw, refreshDur)
	c.JSON(http.StatusOK, sessionResponse(access, refreshRaw, refreshDur, user))
}

// Logout revokes the refresh token and clears cookies.
func (h *AuthHandler) Logout(c *gin.Context) {
	var req refreshRequest
	_ = c.ShouldBindJSON(&req)
	raw := req.RefreshToken
	if raw == "" {
		if ck, err := c.Cookie(refreshCookieName); err == nil {
			raw = ck
		}
	}
	if raw != "" {
		var stored models.RefreshToken
		if err := h.DB.Where("token_hash = ?", utils.HashRefreshToken(raw)).First(&stored).Error; err == nil {
			now := time.Now()
			stored.RevokedAt = &now
			_ = h.DB.Save(&stored).Error
		}
	}
	clearAuthCookies(c, h.Cfg)
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
