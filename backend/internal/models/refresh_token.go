package models

import "time"

// RefreshToken stores a hashed opaque refresh token for rotation.
// The raw token is only returned once at login/refresh; DB keeps the hash.
type RefreshToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	TokenHash string    `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	DeviceInfo string    `json:"device_info,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
