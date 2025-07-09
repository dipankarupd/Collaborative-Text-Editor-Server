package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)
type Document struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	AuthorID  uuid.UUID       `gorm:"type:uuid;not null" json:"author_id"`
	Title     string          `gorm:"not null;default:'Untitled Document'" json:"title"`
	Content   json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"content"`
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

type DocResponse struct {
	ID         uuid.UUID       `json:"id"`
	Author 	   Author			`json:"author"`
	Title      string          `json:"title"`
	Content    json.RawMessage `json:"content"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type Author struct {
	ID   uuid.UUID       `json:"id"`
	Name string          `json:"name"`
}