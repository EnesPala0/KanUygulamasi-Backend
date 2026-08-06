package services

import (
	"fmt"
	"kan-uygulamasi/database"
	"kan-uygulamasi/models"
	"log"
	"time"
)

// StartBloodRequestCleanup, 7 günden eski aktif kan ilanlarını süresi dolmuş (expired) olarak işaretler
// ve kullanıcılara bildirim gönderir. Arka planda sürekli çalışır.
func StartBloodRequestCleanup() {
	// Döngünün ne sıklıkla çalışacağını belirliyoruz (örneğin 12 saatte bir)
	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	// Uygulama başlar başlamaz ilk temizliği hemen yapalım
	performCleanup()

	// Sonra her 12 saatte bir tekrarla
	for {
		<-ticker.C
		performCleanup()
	}
}

func performCleanup() {
	var expiredRequests []models.BloodRequest
	
	// Şu andan 168 saat (7 gün) öncesini bul
	sevenDaysAgo := time.Now().Add(-168 * time.Hour)

	// Status'ü "active" olan VE 7 günden eski olan ilanları çek
	result := database.DB.Where("status = ? AND created_at < ?", "active", sevenDaysAgo).Find(&expiredRequests)
	
	if result.Error != nil {
		log.Println("İlan temizliği (cleanup) sırasında veritabanı hatası:", result.Error)
		return
	}

	if len(expiredRequests) == 0 {
		return // Süresi dolan ilan yok
	}

	fmt.Printf("[CLEANUP] %d adet 7 günden eski ilan bulundu. Kapatılıyor...\n", len(expiredRequests))

	for _, req := range expiredRequests {
		// İlanın durumunu "expired" yap
		req.Status = "expired"
		if err := database.DB.Save(&req).Error; err != nil {
			log.Printf("İlan ID %d güncellenirken hata oluştu: %v\n", req.ID, err)
			continue
		}

		// İlan sahibine bildirim gönder
		notifTitle := "İlan Süresi Doldu"
		notifMessage := "Kan arayışınız 1 haftayı (7 gün) doldurduğu için otomatik olarak kapatıldı. İhtiyacınız devam ediyorsa lütfen yeni bir ilan açın."
		
		err := CreateNotification(req.UserId, "expired", notifTitle, notifMessage, req.ID)
		if err != nil {
			log.Printf("İlan ID %d için bildirim gönderilirken hata: %v\n", req.ID, err)
		}
	}
	
	fmt.Println("[CLEANUP] Otomatik ilan kapatma işlemi tamamlandı.")
}
