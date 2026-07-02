package main

import (
	"fmt"
	"kan-uygulamasi/database"
	"kan-uygulamasi/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Kan Uygulaması — starter")

	//1. burada veritabanı bağlantısını başlatıyoruz
	database.ConnectDB()

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
