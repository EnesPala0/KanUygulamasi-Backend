package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name      string `gorm:"not null" json:"first_name"`
	LastName  string `gorm:"not null" json:"last_name"`
	Email     string `gorm:"not null;unique" json:"email"`
	Password  string `gorm:"not null" json:"password"`
	Phone     string `gorm:"not null" json:"phone"`
	BloodType string `gorm:"not null" json:"blood_type"`
	City      string `gorm:"not null" json:"city"`
	District  string `gorm:"not null" json:"district"`
}
