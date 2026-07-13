# 🩸 Kan Uygulaması - API Dokümantasyonu

Bu doküman, Kan Uygulaması backend servisinin uç noktalarını (endpoints) ve kullanım kurallarını içerir.

**Temel URL:** `http://localhost:8080/api`

---

## 🔓 Herkese Açık Rotalar (Public)
Bu isteklere erişmek için herhangi bir token gerekmez.

| Metot | Uç Nokta (Endpoint) | Açıklama |
| :--- | :--- | :--- |
| **POST** | `/users` | Yeni bir kullanıcı kaydı (Register) oluşturur. |
| **POST** | `/login` | Kullanıcı girişi yapar ve JWT token döndürür. |
| **GET** | `/blood-requests` | Sistemdeki tüm kan taleplerini listeler. |
| **GET** | `/blood-requests/:id` | ID'si verilen spesifik bir kan talebinin detaylarını getirir. |

---

## 🔒 Korumalı Rotalar (Protected)
Bu isteklere erişmek için HTTP başlığında (Header) geçerli bir JWT token gönderilmesi zorunludur.
> **Format:** `Authorization: Bearer <token>`

### 👤 Kullanıcı İşlemleri
| Metot | Uç Nokta (Endpoint) | Açıklama |
| :--- | :--- | :--- |
| **PUT** | `/users/:id` | Kullanıcının profil bilgilerini günceller. |

### 🏥 İlan (Kan Talebi) Yönetimi
| Metot | Uç Nokta (Endpoint) | Açıklama |
| :--- | :--- | :--- |
| **POST** | `/blood-requests` | Yeni bir kan talebi ilanı oluşturur. |
| **PUT** | `/blood-requests/:id` | Mevcut bir kan talebini günceller. |
| **DELETE** | `/blood-requests/:id` | Mevcut bir kan talebini siler. |
| **GET** | `/my-blood-requests` | Giriş yapan kullanıcının kendi açtığı ilanları listeler. |
| **PUT** | `/blood-requests/:id/complete`| İlanı tamamlandı/çözüldü olarak işaretler. |

### 🤝 Gönüllülük ve Başvuru İşlemleri
| Metot | Uç Nokta (Endpoint) | Açıklama |
| :--- | :--- | :--- |
| **POST** | `/volunteers` | Bir kan talebine gönüllü olmak için başvuru yapar. |
| **GET** | `/blood-requests/:id/volunteers` | İlana başvuran gönüllüleri listeler. |
| **GET** | `/my-applications` | Giriş yapan kullanıcının yaptığı başvuruları listeler. |
| **PUT** | `/volunteers/:id/accept` | İlan sahibinin bir gönüllüyü onaylamasını sağlar. |
| **PUT** | `/volunteers/:id/reject` | İlan sahibinin bir gönüllüyü reddetmesini sağlar. |

### 🔔 Bildirim Sistemi
| Metot | Uç Nokta (Endpoint) | Açıklama |
| :--- | :--- | :--- |
| **GET** | `/notifications` | Giriş yapan kullanıcının bildirimlerini listeler. |
| **PUT** | `/notifications/:id/read` | Spesifik bir bildirimi "okundu" olarak işaretler. |