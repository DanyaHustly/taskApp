package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Task struct {
	ID        string         `gorm:"primaryKey" json:"id"`
	Task      string         `json:"task"`
	IsDone    bool           `json:"is_done"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate генерирует uuid если пустой (сохраняет поведение оригинала)
func (t *Task) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
