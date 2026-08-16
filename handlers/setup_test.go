package handlers

import (
	"log"
	"kan-uygulamasi/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB, testler için sqlmock kullanarak sahte bir veritabanı bağlantısı kurar
func setupTestDB() sqlmock.Sqlmock {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("Mock DB oluşturulamadı: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})
	
	// GORM'un default olarak veritabanı versiyonunu sorgulamasını atlıyoruz
	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatalf("GORM mock bağlantısı oluşturulamadı: %v", err)
	}

	// Global database.DB'yi mock DB ile değiştiriyoruz
	database.DB = db

	return mock
}

// setupRouter, HTTP isteklerini test etmek için Gin test ortamını hazırlar
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.Default()
}
