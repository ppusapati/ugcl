package models

import "time"

type OTP struct {
	Id        int32     `db:"id"`
	UserId    int32     `db:"user_id"`
	OtpType   string    `db:"otp_type"`
	OtpValue  int64     `db:"otp_value"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	IsUsed    bool      `db:"is_used"`
}
