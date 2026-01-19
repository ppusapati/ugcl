package models

import "time"

type UserPreferences struct {
	UserID             string            `json:"user_id" db:"user_id"`
	Language           string            `json:"language" db:"language"`
	Timezone           string            `json:"timezone" db:"timezone"`
	DateFormat         string            `json:"date_format" db:"date_format"`
	TimeFormat         string            `json:"time_format" db:"time_format"`
	EmailNotifications bool              `json:"email_notifications" db:"email_notifications"`
	SmsNotifications   bool              `json:"sms_notifications" db:"sms_notifications"`
	PushNotifications  bool              `json:"push_notifications" db:"push_notifications"`
	CustomPreferences  map[string]string `json:"custom_preferences" db:"custom_preferences"`
	CreatedAt          time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at" db:"updated_at"`
}