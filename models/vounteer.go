package models

import "gorm.io/gorm"

type Volunteer struct {
	gorm.Model
	BloodRequestID uint   `gorm:"not null" json:"request_id"`
	UserID         uint   `gorm:"not null" json:"volunteer_id"`
	Status         string `gorm:"not null;default:'pending'" json:"status"`
	// GORM İLİŞKİLERİ:
	User         User          `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user"`
	BloodRequest *BloodRequest `gorm:"foreignKey:BloodRequestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"blood_request,omitempty"`
}
