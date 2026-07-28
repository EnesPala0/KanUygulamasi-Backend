package handlers

import (
	"fmt"
	"kan-uygulamasi/database"
	"kan-uygulamasi/models"
	"kan-uygulamasi/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateBloodRequest, yeni bir kan talebi oluşturmak için HTTP POST isteğini işler.
func CreateBloodRequest(c *gin.Context) {
	var request models.BloodRequest

	// 1. ADIM: Client'tan gelen JSON verisini alıp Go structına çeviriyoruz (Binding).
	if err := c.ShouldBindJSON(&request); err != nil {
		// Eğer JSON formatı bozuksa veya zorunlu (not null) alanlar eksikse, hata döndürüyoruz.
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format or missing required fields", "details": err.Error()})
		return
	}

	// --- İŞTE EKLENEN KRİTİK KISIM ---
	// 1.5 ADIM: Auth Middleware'den gelen kullanıcı ID'sini alıp ilana ekliyoruz
	tokenUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user ID not found in token"})
		return
	}

	// Token'dan gelen ID'yi uint'e çevirip struct'ın UserId alanına atıyoruz
	request.UserId = uint(tokenUserID.(float64))

	var currentUser models.User
	database.DB.Select("latitude", "longitude").Where("id = ?", request.UserId).First(&currentUser)
	request.Latitude = currentUser.Latitude
	request.Longitude = currentUser.Longitude

	// 2.ADIM : Service 'e gidip veritabanına kaydetme işlemini yapıyoruz.
	if err := services.CreateBloodRequest(&request); err != nil {
		// Eğer veritabanına kaydetme sırasında bir hata oluşursa, hata döndürüyoruz.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create blood request", "details": err.Error()})
		return
	}

	//go func() ile işlemi arka plana atıyhoruz böyle API anında yanıt döner
	go func(req models.BloodRequest) {
		var nearbyUsers []models.User

		// GORM ve Earthdistance ile 100.000 metre (100 km) çapındaki kullanıcıları bulma
		err := database.DB.Where("earth_distance(ll_to_earth(latitude, longitude), ll_to_earth(?, ?)) <= ?",
			req.Latitude, req.Longitude, 100000). // İlanın açıldığı koordinat ve 100 km sınırı
			Where("expo_push_token != ''").       // Bildirim token'ı olanları (uygulamaya girenleri) filtrele
			Where("id != ?", req.UserId).         // İlanı açan kişinin KENDİSİNE bildirim gitmesini engelle
			Find(&nearbyUsers).Error

		if err != nil {
			fmt.Printf("Radar araması sırasında hata oluştu: %v\n", err)
			return
		}

		fmt.Printf("RADAR: %s şehrinde açılan ilana 100 km çapında %d adet uygun kullanıcı bulundu!\n", req.City, len(nearbyUsers))

		var tokens []string

		// 1. Bulunan kullanıcıların token'larını bir listeye topluyoruz
		// Aynı zamanda uygulamanın içindeki "Çan" ikonuna (veritabanına) da kaydediyoruz
		for _, user := range nearbyUsers {
			if user.ExpoPushToken != "" {
				tokens = append(tokens, user.ExpoPushToken)
			}

			// (Opsiyonel ama mükemmel detay) Veritabanına da bildirim ekle
			// Böylece adam push bildirimi yanlışlıkla silse bile uygulamaya girince zilde görür
			bildirimMesaji := fmt.Sprintf("%s şehrinde Acil %s kan aranıyor!", req.City, req.RequiredBloodType)
			services.CreateNotification(user.ID, "urgent_need", "Acil Kan İhtiyacı 🩸", bildirimMesaji, req.ID)
		}

		// 2. Eğer listede token varsa, Expo füzelerini ateşle!
		if len(tokens) > 0 {
			baslik := "Acil Kan İhtiyacı!"
			mesaj := fmt.Sprintf("%s civarında %s kana ihtiyaç var. Destek olabilir misin?", req.City, req.RequiredBloodType)

			// Kullanıcı bildirime tıkladığında onu direkt ilana yönlendirmek için gizli veri
			ekstraVeri := map[string]interface{}{
				"blood_request_id": req.ID,
				"type":             "new_request",
			}

			err := services.SendPushNotification(tokens, baslik, mesaj, ekstraVeri)
			if err != nil {
				fmt.Printf("Bildirimler fırlatılamadı: %v\n", err)
			} else {
				fmt.Printf("BİLDİRİM OPERASYONU BAŞARILI! %d telefona mesaj gönderildi.\n", len(tokens))
			}
		}

	}(request)

	// 3.ADIM : Başarılı bir şekilde kaydedildiyse, başarılı mesajı ve oluşturulan kan talebini döndürüyoruz.
	c.JSON(http.StatusCreated, gin.H{
		"message":       "Blood request created successfully",
		"blood_request": request, // oluşan ilanı ID'si ve tarihleri ile birlikte geri döndürüyoruz
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

func GetMyBloodRequests(c *gin.Context) {
	tokenUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user ID not found in token"})
		return
	}

	loggedUserID := uint(tokenUserID.(float64))

	requests, err := services.GetMyBloodRequests(loggedUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve your blood requests", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Successfully retrieved your blood requests",
		"blood_requests": requests,
	})
}

func CompleteBloodRequest(c *gin.Context) {
	requestID := c.Param("id")

	tokenUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user ID not found in token"})
		return
	}

	loggedUserID := uint(tokenUserID.(float64))

	err := services.CompleteBloodRequest(requestID, loggedUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to complete the blood request", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Blood request completed successfully"})

}
