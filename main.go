package main

import (
	"fmt"
	"kan-uygulamasi/database"
	"kan-uygulamasi/routes"
	"kan-uygulamasi/services"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Kan Uygulaması — starter")

	// Sentry initialization
	sentryDsn := os.Getenv("SENTRY_DSN")
	if sentryDsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              sentryDsn,
			EnableTracing:    true,
			TracesSampleRate: 1.0,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
		// Sentry tamponunu program kapanırken temizlemek için
		defer sentry.Flush(2 * time.Second)
		log.Println("Sentry başarıyla başlatıldı.")
	} else {
		log.Println("UYARI: SENTRY_DSN bulunamadı, hata takibi devre dışı.")
	}

	//1. burada veritabanı bağlantısını başlatıyoruz
	database.ConnectDB()

	// Arka planda 7 günlük ilanları otomatik kapatan servisi başlat
	go services.StartBloodRequestCleanup()

	//2. Gin HTTP sunucusunu varsayılan ayarlarla başlatıyoruz
	router := gin.Default()

	//3. Tüm rotaları sunucuya entegre ediyoruz
	routes.SetupRoutes(router)

	//4. sunucuyu 8080 portunda başlatıyoruz
	log.Println("Sunucu başlatılıyor: http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server could not be started: ", err)
	}

}
