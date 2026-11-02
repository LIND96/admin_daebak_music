package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"not null;default:''" json:"-"`
	FullName  string    `json:"full_name"`
	AvatarURL string    `json:"avatar_url"`
	Role      string    `gorm:"not null;default:customer" json:"role"`
	AuthProvider string `gorm:"not null;default:email" json:"auth_provider"`
	GoogleSub    *string `gorm:"uniqueIndex" json:"google_sub,omitempty"`
	// Profil tambahan yang dikumpulkan saat login customer (mobile).
	Gender      string     `json:"gender,omitempty"`
	Birthdate   *time.Time `gorm:"type:date" json:"birthdate,omitempty"`
	Country     string     `json:"country,omitempty"`
	DeviceBrand string     `json:"device_brand,omitempty"`
	DeviceModel string     `json:"device_model,omitempty"`
	OSVersion   string     `json:"os_version,omitempty"`
	Platform    string     `json:"platform,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserResponse is the safe representation returned to clients.
type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	AvatarURL string    `json:"avatar_url"`
	Role      string    `json:"role"`
	AuthProvider string `json:"auth_provider"`
	Gender      string     `json:"gender,omitempty"`
	Birthdate   *time.Time `json:"birthdate,omitempty"`
	Country     string     `json:"country,omitempty"`
	DeviceBrand string     `json:"device_brand,omitempty"`
	DeviceModel string     `json:"device_model,omitempty"`
	OSVersion   string     `json:"os_version,omitempty"`
	Platform    string     `json:"platform,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		FullName:  u.FullName,
		AvatarURL: u.AvatarURL,
		Role:      u.Role,
		AuthProvider: u.AuthProvider,
		Gender:      u.Gender,
		Birthdate:   u.Birthdate,
		Country:     u.Country,
		DeviceBrand: u.DeviceBrand,
		DeviceModel: u.DeviceModel,
		OSVersion:   u.OSVersion,
		Platform:    u.Platform,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}