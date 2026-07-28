package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kan-uygulamasi/database"
	"kan-uygulamasi/models"
	"net/http"
)

func CreateNotification(userID uint, notifType, title, message string, bloodReqID uint) error {
	// 1. Bildirimi veritabanına (uygulama içi zile) kaydet
	notification := models.Notification{
		UserID:         userID,
		Type:           notifType,
		BloodRequestID: bloodReqID,
		Title:          title,
		Message:        message,
	}

	if err := database.DB.Create(&notification).Error; err != nil {
		return err
	}

	// 2. Eğer bu bildirim "urgent_need" (100km radar) ise işlemi burada kes!
	// Neden? Çünkü Handler dosyasında biz zaten bu işlemi "toplu bildirim" (bulk) olarak yapıyoruz.
	// Çift bildirim gitmesini engellemek için bunu atlıyoruz.
	if notifType == "urgent_need" {
		return nil
	}

	// 3. Diğer TÜM bildirimler için (Başvuru, Onay, Red, Tamamlandı) anında telefona füze yolla!
	// Go rutin ile arka plana atıyoruz ki kullanıcıyı bekletmesin.
	go func(uID uint, nType, nTitle, nMsg string, reqID uint) {
		var user models.User

		// Bildirimin gideceği kullanıcının veritabanından Expo Token'ını alıyoruz
		if err := database.DB.Select("expo_push_token").Where("id = ?", uID).First(&user).Error; err == nil {
			if user.ExpoPushToken != "" {
				// Bildirime tıklandığında RN tarafında ilan detayına gitmesi için gizli ID paketi
				extraData := map[string]interface{}{
					"blood_request_id": reqID,
					"type":             nType,
				}

				// Tek kişilik hedefli Expo füzesini ateşle!
				SendPushNotification([]string{user.ExpoPushToken}, nTitle, nMsg, extraData)
			}
		}
	}(userID, notifType, title, message, bloodReqID)

	return nil
}

func GetMyNotifications(userID uint) ([]models.Notification, error) {
	var notifications []models.Notification

	//burada desc ile sıralama yapıyoruz azalan şekilde yani en son gelen bildirim en üstte olacak şekilde
	err := database.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&notifications).Error
	return notifications, err
}

func MarkAsRead(notificationID string, userID uint) error {
	return database.DB.Model(&models.Notification{}).Where("id = ? AND user_id = ?", notificationID, userID).Update("is_read", true).Error

}

// ExpoPushMessage, Expo'nun bizden beklediği standart JSON formatıdır
type ExpoPushMessage struct {
	To    string                 `json:"to"`
	Sound string                 `json:"sound"`
	Title string                 `json:"title"`
	Body  string                 `json:"body"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

// SendPushNotification, elimizdeki token listesine toplu şekilde bildirim fırlatır
func SendPushNotification(tokens []string, title, message string, extraData map[string]interface{}) error {
	// Eğer token listesi boşsa hiç işlem yapmadan çık
	if len(tokens) == 0 {
		return nil
	}

	// 1. Gönderilecek mesaj paketlerini hazırlıyoruz
	var messages []ExpoPushMessage
	for _, token := range tokens {
		// Sadece geçerli, boş olmayan token'ları ekliyoruz
		if token != "" {
			messages = append(messages, ExpoPushMessage{
				To:    token,
				Sound: "default", // Telefondaki varsayılan bildirim sesini çaldırır
				Title: title,
				Body:  message,
				Data:  extraData, // Tıklandığında ilana gitmesi için ilan ID'si gibi veriler
			})
		}
	}

	// Hazırlanan paket kalmadıysa çık
	if len(messages) == 0 {
		return nil
	}

	// 2. Mesajları JSON formatına çeviriyoruz
	jsonBytes, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("JSON oluşturma hatası: %v", err)
	}

	// 3. Expo'nun Bildirim Merkezine HTTP POST isteği oluşturuyoruz
	req, err := http.NewRequest("POST", "https://exp.host/--/api/v2/push/send", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("HTTP istek oluşturma hatası: %v", err)
	}

	// Expo'nun zorunlu tuttuğu başlıklar (Headers)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Content-Type", "application/json")

	// 4. İsteği Ateşle!
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Expo'ya istek atılamadı: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Expo beklenmeyen bir yanıt döndü, Durum Kodu: %d", resp.StatusCode)
	}

	fmt.Printf("Başarılı! %d adet cihaza Expo üzerinden bildirim fırlatıldı.\n", len(messages))
	return nil
}
