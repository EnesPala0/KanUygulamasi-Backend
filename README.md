# 🩸 Kan Bağışı Platformu API (Blood Donation Backend)

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/postgresql-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)
![AWS](https://img.shields.io/badge/AWS-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white)
![Cloudflare](https://img.shields.io/badge/Cloudflare-F38020?style=for-the-badge&logo=Cloudflare&logoColor=white)
![GitHub Actions](https://img.shields.io/badge/github%20actions-%232671E5.svg?style=for-the-badge&logo=githubactions&logoColor=white)

Bu proje, acil kan ihtiyacı olan hastalar ile onlara en yakın konumdaki gönüllü bağışçıları hızlı ve güvenli bir şekilde bir araya getirmeyi amaçlayan, **gerçek zamanlı lokasyon bazlı** bir RESTful API servisidir.

## 🚀 Teknolojik Altyapı & Mimari (Tech Stack)

Sistem, yüksek ölçeklenebilirlik ve güvenlik standartları göz önünde bulundurularak tasarlanmıştır.

* **Backend:** Go (Golang), Gin HTTP Framework
* **Veritabanı:** PostgreSQL (GORM) & **PostGIS/Earthdistance** (Lokasyon hesaplamaları için)
* **Sunucu & Dağıtım:** AWS EC2, Docker & Docker Compose
* **Güvenlik & Ağ:** Cloudflare (WAF, Geo-blocking, SSL/TLS Proxy), JWT, Bcrypt
* **CI/CD Otomasyonu:** GitHub Actions (Sürekli Dağıtım)
* **Harici Servisler:** Resend API (Mail/OTP), Expo Push Notifications (Mobil Bildirim)

---

## ✨ Öne Çıkan Özellikler (Core Features)

* 📍 **Lokasyon Bazlı Akıllı Eşleştirme (Radius Search):** Yeni bir acil kan talebi oluşturulduğunda, PostgreSQL `earthdistance` eklentisi kullanılarak hastanın koordinatlarına en yakın (aciliyet durumuna göre 20km - 500km çapındaki) aktif bağışçılar milisaniyeler içinde tespit edilir.
* 🚀 **Gerçek Zamanlı Push Bildirimleri:** Eşleşen bağışçılara, arka planda çalışan (goroutine) asenkron işlemler sayesinde anında **Expo Push Notification** fırlatılır.
* 🔐 **Güvenlik ve Doğrulama:** 
  * Kullanıcı kayıtlarında **Resend API** üzerinden OTP Mail doğrulaması yapılır.
  * Tüm sistem **JWT (JSON Web Token)** tabanlı yetkilendirme ile korunur.
  * API trafiği **Cloudflare WAF** arkasında gizlenerek yurtdışı bot taramalarına (Geo-blocking) ve DDoS saldırılarına karşı korunur.
* 🤖 **Tam Otomatik CI/CD Pipeline:** Geliştirici `main` dalına kod pushladığı anda, **GitHub Actions** devreye girerek AWS EC2 sunucusuna SSH ile bağlanır, güncel kodları çeker ve Docker konteynerlerini kesintisiz (Zero-Downtime hedefli) şekilde yeniden başlatır.

---

## 🛠️ Kurulum ve Çalıştırma

Proje, tüm bağımlılıkları ile birlikte **Docker** içerisinde çalışacak şekilde yapılandırılmıştır.

1. **Projeyi Klonlayın:**
   ```bash
   git clone https://github.com/EnesPala0/KanUygulamasi-Backend.git
   cd KanUygulamasi-Backend
   ```

2. **Çevre Değişkenlerini (Env) Ayarlayın:**
   Kök dizinde bir `.env` dosyası oluşturup gerekli API key'leri ekleyin (Resend API vb.)

3. **Sistemi Docker ile Ayağa Kaldırın:**
   PostgreSQL veritabanı ve Go API sunucusu tek bir komutla ayağa kalkar:
   ```bash
   docker-compose up --build -d
   ```
   *(API varsayılan olarak `8080` veya `80` portunda çalışacaktır).*

---

## 🔌 API Endpoint'leri

| HTTP Metodu | Endpoint | Açıklama | Yetki |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/users` | Yeni kullanıcı kaydı ve OTP Mail gönderimi | Herkese Açık |
| `POST` | `/api/login` | Giriş yapar ve JWT Token döndürür | Herkese Açık |
| `GET` | `/api/blood-requests` | Filtrelenebilir tüm aktif kan ilanlarını getirir | Herkese Açık |
| `GET` | `/api/blood-requests/:id` | Tek bir ilanın detayını getirir | Herkese Açık |
| `GET` | `/api/users/:id` | Belirli bir kullanıcının (bağışçının) profilini getirir | Herkese Açık |
| `POST` | `/api/users/verify` | OTP kodu doğrulama işlemini yapar | Herkese Açık |
| `POST` | `/api/users/forgot-password` | Şifremi unuttum mailini (OTP) gönderir | Herkese Açık |
| `POST` | `/api/users/reset-password` | Yeni şifre belirleme işlemini yapar | Herkese Açık |
| `POST` | `/api/blood-requests` | **(Asenkron Bildirimli)** Yeni kan ilanı açar | 🔒 JWT Gerekli |
| `PUT` | `/api/blood-requests/:id` | Mevcut ilanı günceller | 🔒 JWT Gerekli |
| `DELETE` | `/api/blood-requests/:id` | İlanı sistemden siler (Soft Delete) | 🔒 JWT Gerekli |
| `PUT` | `/api/blood-requests/:id/complete`| İlanı başarıyla tamamlar ve bağışçılara puan ekler | 🔒 JWT Gerekli |
| `POST` | `/api/volunteers` | İlana gönüllü (bağışçı) başvurusu yapar | 🔒 JWT Gerekli |
| `DELETE` | `/api/volunteers/:id` | Gönüllü başvurusunu iptal eder | 🔒 JWT Gerekli |
| `GET` | `/api/blood-requests/:id/volunteers` | İlana başvuran gönüllüleri listeler (Sadece İlan Sahibi) | 🔒 JWT Gerekli |
| `PUT` | `/api/volunteers/:id/accept` | Gönüllü başvurusunu onaylar | 🔒 JWT Gerekli |
| `PUT` | `/api/volunteers/:id/reject` | Gönüllü başvurusunu reddeder | 🔒 JWT Gerekli |
| `GET` | `/api/my-blood-requests` | Kullanıcının kendi açtığı ilanları getirir | 🔒 JWT Gerekli |
| `GET` | `/api/my-applications` | Kullanıcının yaptığı bağış başvurularını getirir | 🔒 JWT Gerekli |
| `GET` | `/api/notifications` | Kullanıcıya gelen anlık bildirimleri listeler | 🔒 JWT Gerekli |
| `PUT` | `/api/notifications/:id/read` | İlgili bildirimi 'Okundu' olarak işaretler | 🔒 JWT Gerekli |
| `GET` | `/api/me` | Kullanıcının kendi profil bilgilerini getirir | 🔒 JWT Gerekli |
| `PUT` | `/api/users/:id` | Kullanıcının profil bilgilerini günceller | 🔒 JWT Gerekli |
| `PUT` | `/api/user/location` | Kullanıcının anlık koordinatlarını (Enlem/Boylam) günceller | 🔒 JWT Gerekli |
| `PUT` | `/api/users/change-password` | Giriş yapmış kullanıcının şifresini değiştirir | 🔒 JWT Gerekli |
| `DELETE` | `/api/me` | Kullanıcının kendi hesabını siler | 🔒 JWT Gerekli |

---
**Geliştirici:** Enes Pala