package handlers

import (
	"kan-uygulamasi/models"
	"kan-uygulamasi/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateVolunteer(c *gin.Context) {
	var volunteer models.Volunteer
	if err := c.ShouldBindJSON(&volunteer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format or missing required fields", "details": err.Error()})
		return
	}

	tokenUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user ID not found in token"})
		return
	}

	volunteer.UserID = uint(tokenUserID.(float64)) // Token'dan gelen userID'yi Volunteer modeline atıyoruz

	if err := services.CreateVolunteer(&volunteer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create volunteer", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Volunteer created successfully",
		"volunteer": volunteer,
	})

}
