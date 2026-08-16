package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestGenerateOTP(t *testing.T) {
	otp := generateOTP()

	if len(otp) != 6 {
		t.Errorf("KRİTİK HATA: OTP 6 haneli olmalı ancak %d haneli üretildi. Üretilen: %s", len(otp), otp)
	}

	isNumeric := regexp.MustCompile(`^[0-9]+$`).MatchString(otp)
	if !isNumeric {
		t.Errorf("KRİTİK HATA: OTP sadece rakam içermeli ancak farklı karakterler bulundu. Üretilen: %s", otp)
	}

	otp2 := generateOTP()
	if otp == otp2 {
		t.Errorf("GÜVENLİK HATASI: Arka arkaya üretilen iki kod birbiriyle aynı (%s). Rastgelelik bozuk!", otp)
	}
}

func TestLoginUser_InvalidData(t *testing.T) {
	// 1. Test veritabanını ve router'ı hazırlıyoruz
	_ = setupTestDB()
	router := setupRouter()
	router.POST("/login", LoginUser)

	// 2. Eksik veya yanlış formatta bir JSON gönderiyoruz (şifre eksik)
	reqBody := []byte(`{"email": "test@kanbagi.com"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// 3. HTTP İsteğini simüle ediyoruz (httptest recorder kullanarak)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 4. Doğrulama (Assertion): Eksik veri gönderdiğimiz için 400 Bad Request dönmeli
	if w.Code != http.StatusBadRequest {
		t.Errorf("Beklenen HTTP durumu %d ancak %d döndü", http.StatusBadRequest, w.Code)
	}
}

func TestCreateUser_MissingFields(t *testing.T) {
	_ = setupTestDB()
	router := setupRouter()
	router.POST("/users", CreateUser)

	// Eksik veri gönderiyoruz (Kan grubu, Şehir vs. eksik)
	reqBody := []byte(`{
		"first_name": "Ahmet",
		"last_name": "Yılmaz",
		"email": "ahmet@test.com",
		"password": "Password123!"
	}`)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Zorunlu alanlar eksik olduğu için 400 hatası bekliyoruz
	if w.Code != http.StatusBadRequest {
		t.Errorf("Beklenen HTTP durumu %d ancak %d döndü", http.StatusBadRequest, w.Code)
	}
}
