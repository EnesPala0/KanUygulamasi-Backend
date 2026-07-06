package services

import (
	"errors"
	"kan-uygulamasi/database"
	"kan-uygulamasi/models"

	"gorm.io/gorm"
)

func CreateVolunteer(volunteer *models.Volunteer) error {
	var bloodRequest models.BloodRequest
	//ilan gerçekten var mı diye kontrol ediyoruz
	err := database.DB.First(&bloodRequest, volunteer.BloodRequestID).Error
	if err != nil {
		return err
	}
	//ilanın statusu "active" mi diye kontrol ediyoruz
	if bloodRequest.Status != "active" {
		return errors.New("blood request is not active")
	}

	//kullanıcı kendi ilanına başvuramaz
	if bloodRequest.UserId == volunteer.UserID {
		return errors.New("user cannot volunteer for their own blood request")
	}

	//aynı kullanıcı aynı ilana birden fazla başvuramaz
	var existingVolunteer models.Volunteer
	err = database.DB.Where("blood_request_id = ? AND user_id = ?", volunteer.BloodRequestID, volunteer.UserID).First(&existingVolunteer).Error
	if err == nil {
		return errors.New("user has already volunteered for this blood request")
	}

	//eğer dönen hata veritabanında bulamamaktan kaynaklıysa (gorm.ErrRecordNotFound)
	// bu durumda kullanıcıyı oluşturabiliriz
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Eğer başka bir hata oluştuysa, onu döndürüyoruz
		return err
	}

	//başvuru başarılı ise veritabanına kaydediyoruz
	return database.DB.Create(volunteer).Error
}

func GetVolunteersByBloodRequestID(bloodRequestID string) ([]models.Volunteer, error) {
	var volunteers []models.Volunteer
	err := database.DB.Preload("User").Where("blood_request_id = ?", bloodRequestID).Find(&volunteers).Error

	return volunteers, err
}

func AcceptVolunteer(volunteerID string, loggedUserID uint) error {
	var volunteer models.Volunteer

	//1. başvuruyu bul ve ilanı getir
	err := database.DB.Preload("BloodRequest").First(&volunteer, volunteerID).Error
	if err != nil {
		return errors.New("volunteer not found")
	}

	//2. başvurunun ilanının sahibi mi diye kontrol et
	if volunteer.BloodRequest.UserId != loggedUserID {
		return errors.New("you are not authorized to accept this volunteer")
	}

	//3. transaction başlat
	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	//4. başvuru durumunu güncelle
	volunteer.Status = "accepted"
	if err := tx.Save(&volunteer).Error; err != nil {
		tx.Rollback()
		return err
	}

	//5. ilan durumunu güncelliyoruz
	volunteer.BloodRequest.Status = "resolved"
	if err := tx.Save(&volunteer.BloodRequest).Error; err != nil {
		tx.Rollback()
		return err
	}

	//6. islemi onaylıyoruz
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
