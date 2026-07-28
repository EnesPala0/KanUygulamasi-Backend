package models

import (
	"time"
)

type Notification struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `json:"user_id"`
	Type           string    `json:"type"`             // YENİ: Bildirim tipi (örn: urgent_need)
	BloodRequestID uint      `json:"blood_request_id"` // YENİ: Tıklayınca gideceği ilanın ID'si
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}
