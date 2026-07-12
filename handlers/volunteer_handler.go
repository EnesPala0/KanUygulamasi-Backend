package handlers

import (
	"kan-uygulamasi/database"
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

	tokenUserID, exists := c.Get("user_id")
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

func GetVolunteers(c *gin.Context) {
	bloodRequestID := c.Param("id")

	tokenUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user ID not found in token"})
		return
	}

	//tokendan gelen float64 tipli id yi tam sayıya uint e çeviriyoruz
	loggedUserID := uint(tokenUserID.(float64))

	//ilanın sahibi mi diye kontrol ediyoruz
	var bloodRequest models.BloodRequest

	if err := database.DB.First(&bloodRequest, bloodRequestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blood request not found", "details": err.Error()})
		return
	}

	if bloodRequest.UserId != loggedUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view volunteers for this blood request"})
		return
	}

	volunteers, err := services.GetVolunteersByBloodRequestID(bloodRequestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve volunteers", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Volunteers retrieved successfully",
		"volunteers": volunteers,
	})
}

func AcceptVolunteer(c *gin.Context) {
	volunteerID := c.Param("id")

	tokenUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user ID not found in token"})
		return
	}

	loggedUserID := uint(tokenUserID.(float64))

	if err := services.AcceptVolunteer(volunteerID, loggedUserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to accept volunteer", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Volunteer accepted successfully",
	})

}

func GetMyApplications(c *gin.Context) {
	tokenUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user ID not found in token"})
		return
	}

	loggedUserID := uint(tokenUserID.(float64))

	applications, err := services.GetMyApplications(loggedUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve your applications", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Successfully retrieved your applications",
		"applications": applications,
	})
}

func RejectVolunteer(c *gin.Context) {
	volunteerID := c.Param("id")

	tokenUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user ID not found in token"})
		return
	}

	loggedUserID := uint(tokenUserID.(float64))

	if err := services.RejectVolunteer(volunteerID, loggedUserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to reject volunteer", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Volunteer rejected successfully",
	})

}
