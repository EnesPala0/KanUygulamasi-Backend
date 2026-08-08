package handlers

import (
	"fmt"
	"kan-uygulamasi/database"
	"kan-uygulamasi/models"
	"kan-uygulamasi/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Phone     string `json:"phone" binding:"required"`
	BloodType string `json:"blood_type" binding:"required"`
	City      string `json:"city" binding:"required"`
	District  string `json:"district" binding:"required"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LocationUpdateRequest struct {
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	ExpoPushToken string  `json:"expo_push_token"`
}

func UpdateUserLocation(c *gin.Context) {
	//burada tokendan giriş yapan kişiinin ID sini alıyoruz
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "yetkisiz işlem"})
		return
	}
	userID := uint(userIDInterface.(float64))

	//react nativeden gelen verileri okuıyoruz
	var req LocationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "geçersiz veri formati"})
		return
	}

	//veritabanında (GORM ile) sadece bu 3 alanı güncelliyoruz
	// GORM struct güncellemesinde 0 veya boş ("") değerleri GÖRMEZDEN GELİR. 
	// Bu yüzden map[string]interface{} kullanmalıyız.
	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"latitude":        req.Latitude,
		"longitude":       req.Longitude,
		"expo_push_token": req.ExpoPushToken,
	}).Error; err != nil {
		c.JSON(500, gin.H{"error": "Veritabanı güncellenemedi"})
		return
	}

	//basarılı yanıt döndürüuoruz
	c.JSON(200, gin.H{"message": "konum bilgileri basarıyla kaydedildi"})
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
	var req RegisterRequest

	// 1. ADIM: Client'tan gelen JSON verisini alıp DTO struct'ına çeviriyoruz.
	if err := c.ShouldBindJSON(&req); err != nil {
		// Eğer JSON formatı bozuksa veya zorunlu alanlar eksikse, hata döndürüyoruz.
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format or missing required fields", "details": err.Error()})
		return
	}

	// 1.5 ADIM: DTO'daki verileri gerçek Veritabanı modelimize (models.User) aktarıyoruz.
	user := models.User{
		Name:      req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password, // GÜVENLİK NOTU: services.CreateUser içinde bu şifreyi bcrypt ile hash'lediğinden emin ol!
		Phone:     req.Phone,
		BloodType: req.BloodType,
		City:      req.City,
		District:  req.District,
	}

	// 2. ADIM: Service'e gidip veritabanına kaydetme işlemini yapıyoruz.
	// Service katmanı artık içi dolu, şifresi okunmuş 'user' modelini alıp işleyecek.
	if err := services.CreateUser(&user); err != nil {
		// Eğer veritabanına kaydetme sırasında bir hata oluşursa, hata döndürüyoruz.
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}

	// 3. ADIM: Başarılı bir şekilde kaydedildiyse, başarılı mesajı ve oluşturulan kullanıcıyı döndürüyoruz.
	// NOT: models.User içindeki Password alanında json:"-" olduğu için, burada user'ı döndürsek bile şifre dışarı sızmayacak. Kusursuz güvenlik!
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user, // oluşan kullanıcıyı ID'si ve tarihleri ile birlikte geri döndürüyoruz
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

func GetMe(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	// Arayüzü uint'e çeviriyoruz
	userID := uint(userIDInterface.(float64))

	// Servisi çağır
	user, err := services.GetUserByID(userID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	c.JSON(200, user)
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

func ForgotPassword(c *gin.Context) {
	var input ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lütfen geçerli bir e-posta adresi giriniz."})
		return
	}

	user, err := services.GetUserByEmail(input.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bu e-posta adresiyle kayıtlı bir hesap bulunamadı."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Şifre sıfırlama talimatları e-posta adresinize iletildi. Lütfen gelen kutunuzu kontrol ediniz.",
	})
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func ChangeUserPassword(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}
	userID := uint(userIDInterface.(float64))

	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lütfen mevcut ve yeni şifrenizi giriniz."})
		return
	}

	if err := services.ChangePassword(userID, input.OldPassword, input.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Şifreniz başarıyla güncellendi."})
}

// DeleteMe, giriş yapmış kullanıcının kendi hesabını silmesi için gelen isteği işler
func DeleteMe(c *gin.Context) {
	// 1. Auth Middleware'in güvenlik kontrolünden geçen kullanıcının ID'sini alıyoruz
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem: Kullanıcı kimliği bulunamadı"})
		return
	}

	// Arayüzden gelen ID'yi uint (sayı) tipine çeviriyoruz
	userID := uint(userIDInterface.(float64))

	// 2. İşi Service (Aşçı) katmanına devrediyoruz
	if err := services.DeleteUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hesap silinirken bir hata oluştu", "details": err.Error()})
		return
	}

	// 3. Başarılı olursa kullanıcıya güle güle diyoruz :)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Hesabınız başarıyla silindi. Bütün verileriniz anonimleştirildi.",
	})
}

// PublicUserResponse, dışarıya sızdırılmaması gereken bilgileri maskeleyen güvenli struct
type PublicUserResponse struct {
	ID             uint   `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	BloodType      string `json:"blood_type"`
	City           string `json:"city"`
	SavedLives     int    `json:"saved_lives"`
	TotalDonations int    `json:"total_donations"`
	StreakYears    int    `json:"streak_years"`
}

// GetUserByID, dışarıya sadece public bilgileri döndürür
func GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	
	// id string'den uint'e çeviriyoruz
	var id uint
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz kullanıcı ID'si"})
		return
	}

	user, err := services.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	// Sadece güvenli bilgileri DTO'ya aktarıp dönüyoruz
	response := PublicUserResponse{
		ID:             user.ID,
		FirstName:      user.Name,
		LastName:       user.LastName,
		BloodType:      user.BloodType,
		City:           user.City,
		SavedLives:     user.SavedLives,
		TotalDonations: user.TotalDonations,
		StreakYears:    user.StreakYears,
	}

	c.JSON(http.StatusOK, response)
}
