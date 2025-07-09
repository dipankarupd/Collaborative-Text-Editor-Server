package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email" validate:"required,email"`
	Name         string    `gorm:"not null" json:"name" validate:"required,min=2,max=30"`
	PasswordHash *string   `json:"-" validate:"required"` // Always hidden in JSON
	Provider     string    `gorm:"not null;default:'local'" json:"provider" validate:"oneof=local google"`
	CreatedAt    time.Time `gorm:"createdAt" json:"created_at"`
	UpdatedAt    time.Time `gorm:"updatedAt" json:"updated_at"`
}
