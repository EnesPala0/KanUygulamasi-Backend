package models

import "gorm.io/gorm"

type BloodRequest struct {
	gorm.Model
	UserId            uint   `gorm:"not null" json:"user_id"`
	User              User   `gorm:"foreignKey:UserId; constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user"`
	City              string `json:"city"`
	District          string `json:"district"`
	HospitalName      string `gorm:"not null" json:"hospital_name"`
	RequiredBloodType string `gorm:"not null" json:"required_blood_type"`
	RequiredUnits     int    `gorm:"not null" json:"required_units"`
	UrgencyLevel      string `gorm:"not null" json:"urgency_level"`
	Status            string `gorm:"not null; default:'active'" json:"status"`
}
