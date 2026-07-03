package handlers

import (
	"fmt"
	"kan-uygulamasi/models"
	"kan-uygulamasi/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func LoginUser(c *gin.Context) {
	var input LoginInput

	//1. json verisini alıp Go structına çeviriyoruz (Binding).
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format or missing required fields", "details": err.Error()})
		return
	}

	//2. işlemi service katmanına gönderiyoruz
	token, err := services.LoginUser(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password", "details": err.Error()})
		return
	}

	//3. başarılı ise tokenı kullanıcıya gönderiyoz
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}

// CreateUser, yeni bir kullanıcı oluşturmak için HTTP POST isteğini işler.
func CreateUser(c *gin.Context) {
	var user models.User

	//1. ADIM: Client'tan gelen JSON verisini alıp Go structına çeviriyoruz (Binding).
	if err := c.ShouldBindJSON(&user); err != nil {
		//Eğer JSON formatı bozuksa veya zorunlu (not null) alanlar eksikse, hata döndürüyoruz.
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format or missing required fields", "details": err.Error()})
		return
	}

	//2. ADIM : Service 'e gidip veritabanına kaydetme işlemini yapıyoruz.
	if err := services.CreateUser(&user); err != nil {
		//Eğer veritabanına kaydetme sırasında bir hata oluşursa, hata döndürüyoruz.
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}

	//3. ADIM : Başarılı bir şekilde kaydedildiyse, başarılı mesajı ve oluşturulan kullanıcıyı döndürüyoruz.
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user, //oluşan kullanıcıyı IDsi ve tarihleri ile birlikte geri döndürüyoruz
	})
}

func UpdateUser(c *gin.Context) {
	//1. URL'den guncellenmek istenen kullanıcının idsini alalım
	paramID := c.Param("id")

	//2. Middlewareden gelen sisteme giriş yapmış kullanıcıının idsini al
	tokenUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	//JWT içindeki sayılar Goya float64 olarak gelir
	//urlden gelen id ise stringdir. Karşılaştırabilmek için token idsini stringe çeviriyoruz
	tokenIDStr := fmt.Sprintf("%.0f", tokenUserID)

	//3. yetki kontrolu yapıyoruz. Eğer token içindeki kullanıcı ids ile url'den gelen id aynı değilse yetkisiz erişim hatası döndürüyoruz
	if paramID != tokenIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You can only update your own profile"})
		return
	}

	//4. JSON verisini al
	var input services.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format or missing required fields", "details": err.Error()})
		return
	}

	//5. service katmanına gidip veritabanında güncelleme işlemini yapıyoruz
	if err := services.UpdateUserProfile(paramID, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User profile updated successfully"})
}
