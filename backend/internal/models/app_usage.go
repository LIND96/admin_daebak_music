package models

import "time"

// AppUsageEvent records a single anonymous tracking hit from a mobile app.
// Every POST is stored as a new row (no dedup on write). DAU is derived
// at read time via COUNT(DISTINCT device_id). usage_date follows Asia/Jakarta.
// country = full name (e.g. Ukraine), country_code = ISO alpha-2 (e.g. UA).
type AppUsageEvent struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	DeviceID       string    `gorm:"not null;index:idx_usage_device_date" json:"device_id"`
	AppsName       string    `gorm:"not null;index:idx_usage_app_date" json:"apps_name"`
	AndroidVersion string    `json:"android_version,omitempty"`
	DeviceBrand    string    `gorm:"size:100" json:"device_brand,omitempty"`
	DeviceModel    string    `gorm:"size:100" json:"device_model,omitempty"`
	Country        string    `gorm:"size:100" json:"country,omitempty"`
	CountryCode    string    `gorm:"size:10;index" json:"country_code,omitempty"`
	Email          *string   `json:"email,omitempty"`
	OccurredAt     time.Time `gorm:"not null;index" json:"occurred_at"`
	UsageDate      string    `gorm:"type:date;not null;index:idx_usage_device_date;index:idx_usage_app_date" json:"usage_date"`
	CreatedAt      time.Time `json:"created_at"`
}

func (AppUsageEvent) TableName() string {
	return "daily_app_usages"
}
