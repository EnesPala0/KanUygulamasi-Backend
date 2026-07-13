package routes

import (
	"kan-uygulamasi/handlers"
	"kan-uygulamasi/middlewares"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	api := router.Group("/api")

	api.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:5500", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	{
		// --- HERKESE AÇIK ROTALAR (Token gerektirmez) ---
		api.POST("/users", handlers.CreateUser)
		api.POST("/login", handlers.LoginUser) // Fonksiyon ismini 'Login' olarak varsayıyorum

		// İlanları listeleme herkese açık olabilir
		api.GET("/blood-requests", handlers.GetAllBloodRequests)
		api.GET("/blood-requests/:id", handlers.GetBloodRequestByID)

		// --- KORUMALI ROTALAR (Token gerekir) ---
		// Bir alt grup oluşturup AuthMiddleware'i buraya bağlıyoruz
		protected := api.Group("/")
		protected.Use(middlewares.AuthMiddleware())
		{
			// Artık bu rotalar için 'Authorization: Bearer <token>' şart!
			protected.POST("/blood-requests", handlers.CreateBloodRequest)
			protected.PUT("/blood-requests/:id", handlers.UpdateBloodRequest)
			protected.DELETE("/blood-requests/:id", handlers.DeleteBloodRequest)
			protected.PUT("/users/:id", handlers.UpdateUser)
			protected.POST("/volunteers", handlers.CreateVolunteer)
			protected.GET("/blood-requests/:id/volunteers", handlers.GetVolunteers)
			protected.PUT("/volunteers/:id/accept", handlers.AcceptVolunteer)
			protected.GET("/my-blood-requests", handlers.GetMyBloodRequests)
			protected.GET("/my-applications", handlers.GetMyApplications)
			protected.PUT("/volunteers/:id/reject", handlers.RejectVolunteer)
			protected.PUT("/blood-requests/:id/complete", handlers.CompleteBloodRequest)
			protected.GET("/notifications", handlers.GetMyNotifications)
			protected.PUT("/notifications/:id/read", handlers.MarkAsRead)
		}
	}
}
