package handlers

import (
	"kan-uygulamasi/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetMyNotifications(c *gin.Context) {
	tokenUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := uint(tokenUserID.(float64))

	notifications, err := services.GetMyNotifications(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": notifications,
	})
}

func MarkAsRead(c *gin.Context) {
	notificationID := c.Param("id")
	tokenUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := uint(tokenUserID.(float64))

	if err := services.MarkAsRead(notificationID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification could not updated."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification mark as read.",
	})
}
