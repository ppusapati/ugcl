// models/user.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name         string    `gorm:"size:100;not null"`
	Email        string    `gorm:"size:100;uniqueIndex;not null"`
	Phone        string    `gorm:"size:15;uniqueIndex;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	Role         string    `gorm:"size:50;not null;default:'user'"`
	IsActive     bool      `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	return
}
