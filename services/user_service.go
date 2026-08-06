package services

import (
	"errors"
	"fmt"
	"kan-uygulamasi/database"
	"kan-uygulamasi/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UpdateUserInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	City      string `json:"city"`
	District  string `json:"district"`
	BloodType string `json:"blood_type"`
}

func CreateUser(user *models.User) error {
	var existingUser models.User

	// veritabanınıda aynı email ile kayıtlı kullanıcı var mı diye kontrol ediyoruz
	err := database.DB.Where("email = ?", user.Email).First(&existingUser).Error

	if err == nil {
		// Eğer aynı email ile kayıtlı kullanıcı varsa, hata döndürüyoruz
		return fmt.Errorf("user with email %s already exists", user.Email)
	}
	//eğer dönen hata veritabanında bulamamaktan kaynaklıysa (gorm.ErrRecordNotFound)
	// bu durumda kullanıcıyı oluşturabiliriz
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Eğer başka bir hata oluştuysa, onu döndürüyoruz
		return err
	}

	//kullanıcının düz metin şifresini alıp geri döndürülemez bir hashe çeviriyoruz
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Failed to hash password: %v", err)
	}

	//modelin içindeki şifreyi haslenmiş versiyon ile değiştirdik
	user.Password = string(hashedPassword)

	return database.DB.Create(user).Error
}

func LoginUser(email, password string) (string, error) {
	var user models.User

	//kullanıcıyı email ile veritabanında arıyoruz

	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return "", errors.New("invalid email or password")
	}

	//gelen düz şifre ile veritabanındaki haslenmiş şifreyi karşılaştırırız
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	//şifre doğru ise JWT token oluşturuyoruz
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,                               //token içine kullanıcı idsini gömüyoruz
		"exp":     time.Now().Add(time.Hour * 72).Unix(), //token 72 saat geçerli olacak
	})

	//tokenı .env içindeki secret key ile imzalıyoruz
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET"))) //secret key
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return tokenString, nil
}

func UpdateUserProfile(userID string, input UpdateUserInput) error {
	var user models.User

	//1. kullanıcıyı veritabanında buluyoruz
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	//2. gelen verilerle kullanıcının bilgilerini güncelle
	updates := map[string]interface{}{}
	if input.FirstName != "" {
		updates["name"] = input.FirstName
	}
	if input.LastName != "" {
		updates["last_name"] = input.LastName
	}
	if input.Phone != "" {
		updates["phone"] = input.Phone
	}
	if input.City != "" {
		updates["city"] = input.City
	}
	if input.District != "" {
		updates["district"] = input.District
	}
	if input.BloodType != "" {
		updates["blood_type"] = input.BloodType
	}

	err := database.DB.Model(&user).Updates(updates).Error
	return err
}

func GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("kullanıcı bulunamadı")
	}

	// Eski şifre kontrolü
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("mevcut şifreniz hatalı")
	}

	// Yeni şifreyi hashle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("şifre hashlenirken hata oluştu: %v", err)
	}

	return database.DB.Model(&user).Update("password", string(hashedPassword)).Error
}

func DeleteUser(userID uint) error {
	var user models.User

	// 1. Kullanıcıyı bul
	if err := database.DB.First(&user, userID).Error; err != nil {
		return err
	}

	// 2. Kullanıcının kişisel verilerini temizle (Anonimleştirme)
	// Böylece App Store ve KVKK kurallarına tam uymuş oluruz.
	database.DB.Model(&user).Updates(map[string]interface{}{
		"name":            "Silinmiş",
		"last_name":       "Kullanıcı",
		"email":           fmt.Sprintf("deleted_%d@kanuygulamasi.com", user.ID),
		"phone":           "0000000000",
		"expo_push_token": "", // Bildirimleri tamamen kapatıyoruz
		"latitude":        0,
		"longitude":       0,
	})

	// 3. Kullanıcının yaptığı tüm gönüllü başvurularını (Volunteer) temizle
	database.DB.Where("user_id = ?", user.ID).Delete(&models.Volunteer{})

	// 4. Kullanıcının açtığı tüm ilanları bul
	var userRequests []models.BloodRequest
	if err := database.DB.Where("user_id = ?", user.ID).Find(&userRequests).Error; err == nil {
		for _, req := range userRequests {
			// İlana yapılan tüm gönüllü başvurularını sil
			database.DB.Where("blood_request_id = ?", req.ID).Delete(&models.Volunteer{})
			// İlanın kendisini sil
			database.DB.Delete(&req)
		}
	}

	// 5. GORM ile Soft Delete İşlemi (deleted_at alanına şu anki tarihi yazar)
	if err := database.DB.Delete(&user).Error; err != nil {
		return err
	}

	return nil
}
