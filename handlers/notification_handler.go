package handlers

import (
	"kan-uygulamasi/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetMyNotifications(c *gin.Context) {
	tokenUserID, exists := c.Get("userID")
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
