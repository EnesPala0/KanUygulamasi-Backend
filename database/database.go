package database

import (
	"fmt"
	"kan-uygulamasi/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB, tüm projede kullanacağımız global veritabanı değişkeni
var DB *gorm.DB

func ConnectDB() {
	// DSN (Data Source Name) Veritabanı bağlantı bilgilerimiz
	//burda .env dosyasını yüklüyoruz önce error varsa hata
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//os..Getenv ile .env dosyasındaki değişkenleri alıyoruz
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	//değişkenleri dinamik olarak string içine gömüyoruz
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	//bağlantı başarılı mı kontrol ediyoruz
	if err != nil {
		log.Fatal("Failed to connect to database! Error: \n", err)
	}
	log.Println("Database connection established successfully!")

	//tabloları veritabanına otomatik olarak yansıtıyoruz
	log.Println("Models are being migrated...")

	err = DB.AutoMigrate(&models.User{}, &models.BloodRequest{})
	if err != nil {
		log.Fatal("Failed to migrate models! Error: \n", err)
	}
	log.Println("Models migrated successfully!")

}
