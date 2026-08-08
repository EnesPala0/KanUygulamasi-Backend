package middlewares

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Global limiter: Saniyede 20 istek, maksimum 50 burst kapasiteli
var limiter = rate.NewLimiter(rate.Every(time.Second/20), 50)

// RateLimitMiddleware, sisteme gelen isteklerin hızını kısıtlayarak DoS saldırılarını engeller
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Çok fazla istek gönderildi. Lütfen biraz bekleyip tekrar deneyin.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
