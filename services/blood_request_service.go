package services

import (
	"errors"
	"kan-uygulamasi/database"
	"kan-uygulamasi/models"
)

type BloodRequestFilter struct {
	City      string `form:"city"`
	District  string `form:"district"`
	BloodType string `form:"blood_type"`
	Urgency   string `form:"urgency_level"`
}

// CreateBloodRequest, yeni bir kan talebi oluşturur ve veritabanına kaydeder.
func CreateBloodRequest(bloodRequest *models.BloodRequest) error {

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

	// Sorguyu çalıştır ve sonuçları requests dizisine aktar
	err := query.Find(&requests).Error

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

	//gelen verilerle eski ilanı güncelliyoruz
	err = database.DB.Model(&currentRequest).Select(
		"City",
		"District",
		"HospitalName",
		"RequiredBloodType",
		"RequiredUnits",
		"UrgencyLevel",
	).Updates(updatedData).Error

	if err != nil {
		return currentRequest, err
	}
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

	//3. ilan zaten kapalıysa boşuna işlem yapma
	if request.Status == "resolved" {
		return errors.New("request is already completed")
	}

	//4. ilanı kapatıyoruz
	request.Status = "resolved"
	if err := database.DB.Save(&request).Error; err != nil {
		return err
	}

	return nil
}
