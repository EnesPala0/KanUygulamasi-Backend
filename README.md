# 🩸 Kan Bağışı API (Blood Donation API)

Bu proje, acil kan ihtiyacı olan hastalar ile gönüllü bağışçıları hızlı ve güvenli bir şekilde bir araya getirmeyi amaçlayan, modern mimariyle tasarlanmış bir RESTful API servisidir.

## 🚀 Teknolojik Altyapı (Tech Stack)

* **Programlama Dili:** Go (Golang)
* **Web Framework:** Gin HTTP Framework
* **Veritabanı & ORM:** PostgreSQL, GORM
* **Güvenlik & Kimlik Doğrulama:** JWT (JSON Web Token), Bcrypt
* **Altyapı & Dağıtım:** Docker, Docker Compose

## ✨ Temel Özellikler

* **Güvenli Kullanıcı Yönetimi:** Kullanıcı şifreleri veritabanına doğrudan yazılmaz, Bcrypt algoritması ile hash'lenerek üst düzey güvenlik sağlanır.
* **JWT Entegrasyonu (Auth Middleware):** Kullanıcı giriş işlemlerinde (Login) benzersiz bir token üretilir. İlan ekleme, silme ve güncelleme gibi kritik işlemler sadece doğrulanmış token'a sahip kullanıcılar tarafından yapılabilir.
* **İlan Yönetimi (CRUD):** Kullanıcılar aciliyet durumu, kan grubu, şehir ve hastane gibi detayları belirterek kan bağışı talepleri oluşturabilir.
* **Konteynerizasyon:** Docker Compose sayesinde sistem herhangi bir bilgisayarda dışa bağımlılık gerektirmeden tek bir komutla ayağa kaldırılabilir.

## 🛠️ Kurulum ve Çalıştırma

Projeyi yerel ortamınızda çalıştırmak için bilgisayarınızda **Docker** ve **Docker Compose** kurulu olmalıdır.

1. **Projeyi Klonlayın:**
   ```bash
   git clone [https://github.com/KULLANICI_ADIN/KanUygulamasi-Backend.git](https://github.com/KULLANICI_ADIN/KanUygulamasi-Backend.git)
   cd KanUygulamasi-Backend

2. **Sistemi Ayağa Kaldırın:**
    PostgreSQL veritabanını ve gerekli tüm altyapıyı Docker üzerinden başlatmak için:
    ```bash
    docker-compose up -d

    Uygulamayı Çalıştırın:
    ```bash
    go run .   

HTTP Metodu,Endpoint,Açıklama,Yetki Gereksinimi
POST,/api/users,Yeni kullanıcı kaydı oluşturur,Herkese Açık
POST,/api/login,Giriş yapar ve JWT Token döndürür,Herkese Açık
GET,/api/blood-requests,Tüm kan ilanlarını listeler,Herkese Açık
GET,/api/blood-requests/:id,Tek bir ilanın detaylarını getirir,Herkese Açık
POST,/api/blood-requests,Yeni bir kan ilanı oluşturur,🔒 Bearer Token
PUT,/api/blood-requests/:id,Mevcut ilanı günceller,🔒 Bearer Token
DELETE,/api/blood-requests/:id,İlanı sistemden siler,🔒 Bearer Token

Geliştirici: Enes Pala