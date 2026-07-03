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
	//GORM un Updates metodu ile sadece içi dolu olan alanları güncelleyebiliriz. Boş alanlar güncellenmez.
	err := database.DB.Model(&user).Updates(models.User{
		Phone:     input.Phone,
		City:      input.City,
		District:  input.District,
		BloodType: input.BloodType,
	}).Error

	return err
}
