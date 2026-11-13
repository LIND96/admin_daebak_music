package models

import "time"

// ChatMessage records a single AI-chat relay hit from a mobile app.
// One row per POST (success or provider failure alike) for monitoring and
// per-app/per-idol statistics. usage_date follows Asia/Jakarta, same
// convention as daily_app_usages.
type ChatMessage struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DeviceID   string    `gorm:"not null;index:idx_chat_device_date" json:"device_id"`
	AppsName   string    `gorm:"not null;index:idx_chat_app_date" json:"apps_name"`
	IdolName   string    `gorm:"size:100" json:"idol_name,omitempty"`
	Provider   string    `gorm:"size:128" json:"provider,omitempty"`
	TokensIn   int       `json:"tokens_in,omitempty"`
	TokensOut  int       `json:"tokens_out,omitempty"`
	Ok         bool      `gorm:"not null" json:"ok"`
	OccurredAt time.Time `gorm:"not null;index" json:"occurred_at"`
	UsageDate  string    `gorm:"type:date;not null;index:idx_chat_device_date;index:idx_chat_app_date" json:"usage_date"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
