package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name           string  `gorm:"not null" json:"first_name"`
	LastName       string  `gorm:"not null" json:"last_name"`
	Email          string  `gorm:"not null;unique" json:"email"`
	Password       string  `gorm:"not null" json:"-"`
	Phone          string  `gorm:"not null" json:"phone"`
	BloodType      string  `gorm:"not null" json:"blood_type"`
	City           string  `gorm:"not null" json:"city"`
	District       string  `gorm:"not null" json:"district"`
	TotalDonations int     `gorm:"default:0" json:"total_donations"`
	SavedLives     int     `gorm:"default:0" json:"saved_lives"`
	StreakYears    int     `gorm:"default:0" json:"streak_years"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	ExpoPushToken  string  `json:"expo_push_token"`
}
