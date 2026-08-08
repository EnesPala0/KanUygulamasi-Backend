package services

import (
	"errors"
	"kan-uygulamasi/database"
	"kan-uygulamasi/models"

	"gorm.io/gorm"
)

type BloodRequestFilter struct {
	City      string `form:"city"`
	District  string `form:"district"`
	BloodType string `form:"blood_type"`
	Urgency   string `form:"urgency_level"`
}

// CreateBloodRequest, yeni bir kan talebi oluşturur ve veritabanına kaydeder.
func CreateBloodRequest(bloodRequest *models.BloodRequest) error {

	// Sanity Check: İstenen ünite sıfır veya negatif olamaz
	if bloodRequest.RequiredUnits <= 0 {
		bloodRequest.RequiredUnits = 1
	}

	//buraya ileride kurallar eklenecek, örneğin aynı kullanıcıdan aynı kan grubunda birden fazla talep oluşturulmasını engellemek gibi.
	//Şimdilik veritabanına kaydetme işlemi yapıyoruz
	err := database.DB.Create(bloodRequest).Error
	return err
}

// GetAllBloodRequests, tek tek string yerine filtre paketimizi struct olarak alıyoruz
func GetAllBloodRequests(filter BloodRequestFilter) ([]models.BloodRequest, error) {
	var requests []models.BloodRequest

	// Preload("User") sayesinde, her ilanın içine o ilanı açan kullanıcının bilgilerini gömüyoruz!
	query := database.DB.Preload("User")

	//---DİNAMİK SORGULAR---
	//eğeer structın içindeki dğerler boş değilse o filtreyi WHERE şartı olarak SQL e ekliyoruz
	if filter.City != "" {
		query = query.Where("city = ?", filter.City)
	}
	if filter.District != "" {
		query = query.Where("district = ?", filter.District)
	}
	if filter.BloodType != "" {
		query = query.Where("required_blood_type = ?", filter.BloodType)
	}
	if filter.Urgency != "" {
		query = query.Where("urgency_level = ?", filter.Urgency)
	}

	// Sorguyu çalıştır ve sonuçları requests dizisine aktar (Maksimum 50 kayıt getir ve en yeni en üstte olsun)
	err := query.Order("created_at desc").Limit(50).Find(&requests).Error

	return requests, err
}

// GetBloodRequestByID, ilan ID'sine göre ilgili kan talebini getirir.
func GetBloodRequestByID(id string) (models.BloodRequest, error) {
	var request models.BloodRequest

	//find yerine First() kullanıyoruz çünkü ID tekil bir değer ve sadece bir kayıt dönecek.
	// GORM, First(&request, id) dediğimizde arka planda "WHERE id = ?" sorgusunu kendi yazar.
	err := database.DB.Preload("User").First(&request, id).Error

	return request, err
}

func UpdateBloodRequest(id string, updatedData *models.BloodRequest) (models.BloodRequest, error) {
	var currentRequest models.BloodRequest

	//veritabanından eski ilanı çekiyoruz
	err := database.DB.First(&currentRequest, id).Error
	if err != nil {
		return currentRequest, err
	}

	// Sadece dolu gelen (boş/sıfır olmayan) alanları güncelleyelim ki diğer alanlar sıfırlanmasın/bozulmasın
	updates := make(map[string]interface{})
	if updatedData.City != "" {
		updates["city"] = updatedData.City
	}
	if updatedData.District != "" {
		updates["district"] = updatedData.District
	}
	if updatedData.HospitalName != "" {
		updates["hospital_name"] = updatedData.HospitalName
	}
	if updatedData.RequiredBloodType != "" {
		updates["required_blood_type"] = updatedData.RequiredBloodType
	}
	if updatedData.RequiredUnits != 0 {
		updates["required_units"] = updatedData.RequiredUnits
	}
	if updatedData.UrgencyLevel != "" {
		updates["urgency_level"] = updatedData.UrgencyLevel
	}
	if updatedData.Status != "" {
		updates["status"] = updatedData.Status
	}
	if updatedData.MedicalNote != "" {
		updates["medical_note"] = updatedData.MedicalNote
	}

	err = database.DB.Model(&currentRequest).Updates(updates).Error
	if err != nil {
		return currentRequest, err
	}

	// En güncel halini tekrar veritabanından çekip dönelim
	database.DB.Preload("User").First(&currentRequest, id)
	return currentRequest, nil
}

// soft-delete yöntemi ile sileceğiz
func DeleteBloodRequest(id string) error {
	var request models.BloodRequest

	//veritabanından ilanı çekiyoruz
	err := database.DB.First(&request, id).Error
	if err != nil {
		return err
	}

	// İlana bağlı olan tüm gönüllü başvurularını da temizle (Cascade soft-delete)
	database.DB.Where("blood_request_id = ?", request.ID).Delete(&models.Volunteer{})

	err = database.DB.Delete(&request).Error
	return err
}

func GetMyBloodRequests(userID uint) ([]models.BloodRequest, error) {
	var requests []models.BloodRequest
	err := database.DB.Preload("User").Where("user_id = ?", userID).Find(&requests).Error
	return requests, err
}

func CompleteBloodRequest(requestID string, loggedUserID uint) error {
	var request models.BloodRequest

	//1. ilanı veritabanından buuluyoruz
	if err := database.DB.First(&request, requestID).Error; err != nil {
		return err
	}

	//2. ilanın sahibi ile giriş yapan kullanıcının ID'sini karşılaştırıyoruz
	if request.UserId != loggedUserID {
		return errors.New("unauthorized: only the owner can complete the request")
	}

	//3. ilan zaten kapalıysa boşuna hata dönme, başarılı say (idempotent)
	if request.Status == "resolved" || request.Status == "Tamamlandı" || request.Status == "completed" {
		return nil
	}

	//4. ilanı kapatıyoruz
	if err := database.DB.Model(&request).Update("status", "resolved").Error; err != nil {
		return err
	}

	//5. Bu ilana başvurmuş ve "approved" (Onaylandı) durumunda olan gönüllüleri bulup sayaçlarını artırıyoruz!
	var approvedVolunteers []models.Volunteer
	if err := database.DB.Where("blood_request_id = ? AND status IN ?", request.ID, []string{"approved", "accepted", "onaylandı", "kabul"}).Find(&approvedVolunteers).Error; err == nil {
		for _, vol := range approvedVolunteers {
			// Bir bağışçı genelde 1 ünite kan verir ve 1 ünite kan 3 hayat kurtarır (Kızılay standardı).
			// Her gönüllüye 1 bağış ve 3 kurtarılan hayat ekliyoruz.
			database.DB.Model(&models.User{}).Where("id = ?", vol.UserID).
				Updates(map[string]interface{}{
					"saved_lives":     gorm.Expr("saved_lives + ?", 3),
					"total_donations": gorm.Expr("total_donations + ?", 1),
					"streak_years":    gorm.Expr("CASE WHEN streak_years = 0 THEN 1 ELSE streak_years END"),
				})
			CreateNotification(
				vol.UserID,
				"completed",
				"Harika Bir İş Başardınız!",
				"Destek olduğunuz kan talebi başarıyla tamamlandı. Yaptığınız 1 ünite kan bağışı ile 3 hayat kurtardınız! Sonsuz teşekkürler. ❤️",
				request.ID,
			)
		}
	}

	return nil
}
