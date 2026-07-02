package services

import (
	"kan-uygulamasi/database"
	"kan-uygulamasi/models"
)

// CreateBloodRequest, yeni bir kan talebi oluşturur ve veritabanına kaydeder.
func CreateBloodRequest(bloodRequest *models.BloodRequest) error {

	//buraya ileride kurallar eklenecek, örneğin aynı kullanıcıdan aynı kan grubunda birden fazla talep oluşturulmasını engellemek gibi.
	//Şimdilik veritabanına kaydetme işlemi yapıyoruz
	err := database.DB.Create(bloodRequest).Error
	return err
}

// GetAllBloodRequests, gelen şehir parametresine göre kan taleplerini listeler veya tümünü getirir.
func GetAllBloodRequests(city string) ([]models.BloodRequest, error) {
	var requests []models.BloodRequest

	// Preload("User") sayesinde, her ilanın içine o ilanı açan kullanıcının bilgilerini gömüyoruz!
	query := database.DB.Preload("User")

	// Eğer garson bize bir şehir gönderdiyse (parametre boş değilse) filtreyi uygula
	if city != "" {
		query = query.Where("city = ?", city)
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
	err = database.DB.Model(&currentRequest).Updates(updatedData).Error
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
