package handlers

import (
	"kan-uygulamasi/models"
	"kan-uygulamasi/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateBloodRequest, yeni bir kan talebi oluşturmak için HTTP POST isteğini işler.
func CreateBloodRequest(c *gin.Context) {
	var request models.BloodRequest

	//1. ADIM: Client'tan gelen JSON verisini alıp Go structına çeviriyoruz (Binding).
	if err := c.ShouldBindJSON(&request); err != nil {
		//Eğer JSON formatı bozuksa veya zorunlu (not null) alanlar eksikse, hata döndürüyoruz.
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format or missing required fields", "details": err.Error()})
		return
	}

	//2.ADIM : Service 'e gidip veritabanına kaydetme işlemini yapıyoruz.
	if err := services.CreateBloodRequest(&request); err != nil {
		//Eğer veritabanına kaydetme sırasında bir hata oluşursa, hata döndürüyoruz.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create blood request", "details": err.Error()})
		return
	}

	//3.ADIM : Başarılı bir şekilde kaydedildiyse, başarılı mesajı ve oluşturulan kan talebini döndürüyoruz.
	c.JSON(http.StatusCreated, gin.H{
		"message":       "Blood request created successfully",
		"blood_request": request, //oluşan ilanı IDsi ve tarihleri ile birlikte geri döndürüyoruz
	})
}

func GetAllBloodRequests(c *gin.Context) {
	var filter services.BloodRequestFilter

	//urldeki ?city=...&district=...&blood_type=...&urgency_level=... parametrelerini alıyoruz
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters", "details": err.Error()})
		return
	}

	//doldurulmuş filtreyi service katmanına gönderiyoruz
	requests, err := services.GetAllBloodRequests(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve blood requests", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"blood_requests": requests})

}

// İlanın detay sayfasına tıklandıgında, ilan ID'si ile birlikte bu fonksiyon çağrılır ve ilgili ilanı getirir.
func GetBloodRequestByID(c *gin.Context) {
	//c.Query() "*?" işaretinden sonrasını alırdı
	//c.Param() ise "/blood_requests/:id" kısmındaki ":id" değerini alır.
	id := c.Param("id")

	request, err := services.GetBloodRequestByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retrieve blood request or it is deleted", "details": err.Error()})
		return
	}

	//ilan başarılı bir şekilde bulunduysa, ilan detaylarını JSON formatında döndürüyoruz.
	c.JSON(http.StatusOK, gin.H{"blood_request": request})
}

func UpdateBloodRequest(c *gin.Context) {
	id := c.Param("id")

	var updatedData models.BloodRequest
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format or missing required fields", "details": err.Error()})
		return
	}

	updatedRequest, err := services.UpdateBloodRequest(id, &updatedData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update blood request", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"blood_request": updatedRequest})
}

func DeleteBloodRequest(c *gin.Context) {
	id := c.Param("id")

	err := services.DeleteBloodRequest(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "The blood request could not be found or has already been deleted.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "The blood request was deleted successfully (Soft Delete applied).",
	})
}
