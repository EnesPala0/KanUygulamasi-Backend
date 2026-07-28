package services

import (
	"errors"
	"kan-uygulamasi/database"
	"kan-uygulamasi/models"
	"strings"

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
	statusLower := strings.ToLower(strings.TrimSpace(bloodRequest.Status))
	if statusLower != "active" && statusLower != "aktif" && statusLower != "açık" && statusLower != "yayında" && statusLower != "" {
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
	if err := database.DB.Create(volunteer).Error; err != nil {
		return err
	}

	CreateNotification(
		bloodRequest.UserId,
		"volunteer_applied", // YENİ: Tipi belirttik, RN'de person-add ikonu çıkacak
		"New Volunteer Application",
		"You have a new volunteer application for your blood request. Check it out!",
		bloodRequest.ID, // YENİ: İlan ID'sini gönderdik
	)

	return nil

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
	if err := tx.Model(&volunteer).Update("status", "accepted").Error; err != nil {
		tx.Rollback()
		return err
	}

	//5. ilan durumunu güncelliyoruz
	if err := tx.Model(volunteer.BloodRequest).Update("status", "resolved").Error; err != nil {
		tx.Rollback()
		return err
	}

	//6. islemi onaylıyoruz
	if err := tx.Commit().Error; err != nil {
		return err
	}

	CreateNotification(
		volunteer.UserID,
		"application_approved", // YENİ: Onaylandı tipi
		"Your volunteer Application Accepted",
		"Thank you for your blood support!. Please contact the blood request owner for further details.",
		volunteer.BloodRequestID, // YENİ
	)
	return nil
}

func GetMyApplications(userID uint) ([]models.Volunteer, error) {
	var applications []models.Volunteer

	err := database.DB.Preload("BloodRequest").Where("user_id = ?", userID).Find(&applications).Error
	return applications, err
}

func RejectVolunteer(volunteerID string, loggedUserID uint) error {
	var volunteer models.Volunteer
	//1. başvuruyu bul ve ilanı getir
	if err := database.DB.Preload("BloodRequest").First(&volunteer, volunteerID).Error; err != nil {
		return errors.New("application not found")
	}

	if volunteer.BloodRequest.UserId != loggedUserID {
		return errors.New("you are not authorized to reject this application")
	}

	if err := database.DB.Model(&volunteer).Update("status", "rejected").Error; err != nil {
		return errors.New("failed to reject the application")
	}

	CreateNotification(
		volunteer.UserID,
		"application_rejected", // YENİ: Reddedildi tipi
		"Your volunteer Application Rejected",
		"Unfortunately, your volunteer application has been rejected. We appreciate your willingness to help.",
		volunteer.BloodRequestID, // YENİ
	)

	return nil
}

func DeleteVolunteer(volunteerID string, loggedUserID uint) error {
	var volunteer models.Volunteer
	if err := database.DB.First(&volunteer, volunteerID).Error; err != nil {
		return errors.New("volunteer application not found")
	}
	if volunteer.UserID != loggedUserID {
		return errors.New("you can only cancel your own applications")
	}
	return database.DB.Unscoped().Delete(&volunteer).Error
}

func DeleteVolunteerByRequestID(requestID string, loggedUserID uint) error {
	var volunteer models.Volunteer
	if err := database.DB.Where("blood_request_id = ? AND user_id = ?", requestID, loggedUserID).First(&volunteer).Error; err != nil {
		return errors.New("volunteer application not found")
	}
	return database.DB.Unscoped().Delete(&volunteer).Error
}
