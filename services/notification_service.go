package services

import (
	"kan-uygulamasi/database"
	"kan-uygulamasi/models"
)

func CreateNotification(userID uint, title, message string) error {
	notification := models.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
	}

	return database.DB.Create(&notification).Error
}

func GetMyNotifications(userID uint) ([]models.Notification, error) {
	var notifications []models.Notification

	err := database.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&notifications).Error
	return notifications, err
}

func MarkAsRead(notificationID string, userID uint) error {
	return database.DB.Model(&models.Notification{}).Where("id = ? AND user_id = ?", notificationID, userID).Update("is_read", true).Error

}
