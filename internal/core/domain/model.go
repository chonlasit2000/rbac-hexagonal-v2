package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	Seq       int64          `gorm:"index"`
	Uid       uuid.UUID      `gorm:"primaryKey;type:uuid"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Hook สร้าง UUID และ Seq อัตโนมัติ
func (m *Model) BeforeCreate(tx *gorm.DB) (err error) {
	if m.Uid == uuid.Nil {
		m.Uid = uuid.New()
	}
	m.Seq = time.Now().UnixNano()
	return
}
