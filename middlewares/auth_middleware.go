package middlewares

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware, korumalı rotalara gelen isteklerdeki JWT token'ı doğrular
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. İsteğin "Authorization" header'ına bak
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Erişim reddedildi: Token bulunamadı"})
			c.Abort() // İsteği burada keser, ileriye göndermez
			return
		}

		// 2. Token formatını kontrol et (Genelde "Bearer <token>" şeklinde gelir)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz token formatı"})
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 3. Token'ı gizli anahtarımızla çöz ve doğrula
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("beklenmeyen imza metodu")
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// 4. Hata varsa veya token geçersizse kapıdan çevir
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz veya süresi dolmuş token"})
			c.Abort()
			return
		}

		// 5. Her şey yolundaysa, token'ın içindeki user_id'yi alıp işlemin devamına aktar
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Bu sayede ilan oluştururken "hangi kullanıcı oluşturdu" bilgisini buradan çekebileceğiz
			c.Set("user_id", claims["user_id"])
			c.Next() // Güvenlikten geçti, asıl fonksiyona (Handler'a) devam edebilir
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz erişim"})
			c.Abort()
			return
		}
	}
}
