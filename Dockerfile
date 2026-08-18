# ==========================================
# AŞAMA 1: Builder (Derleme Aşaması)
# ==========================================
# İçinde Go kurulu olan ağır bir Linux kutusu alıyoruz
FROM golang:alpine AS builder

# Çalışma dizinini belirliyoruz
WORKDIR /app

# Önce bağımlılık dosyalarını kopyalayıp indiriyoruz (Cache avantajı sağlar)
COPY go.mod go.sum ./
RUN go mod download

# Tüm kodumuzu kopyalıyoruz
COPY . .

# Uygulamayı Linux'a uygun olarak (.exe olmadan) tek parça derliyoruz
RUN CGO_ENABLED=0 GOOS=linux go build -o kan-uygulamasi main.go

# ==========================================
# AŞAMA 2: Final (Çalıştırma Aşaması)
# ==========================================
# Sadece 5 MB olan içi bomboş, güvenli bir Alpine Linux kutusu alıyoruz
FROM alpine:latest

WORKDIR /root/

# Saat dilimi ayarları (Logların Türkiye saatine uygun görünmesi için faydalıdır)
RUN apk --no-cache add tzdata
ENV TZ=Europe/Istanbul

# Aşama 1'den (builder) SADECE derlenmiş tek dosyayı kopyalıyoruz! 
# (Go kodlarımız, gereksiz SDK'lar vs. hiçbiri bu kutuya girmiyor)
COPY --from=builder /app/kan-uygulamasi .

# Uygulamamızın 8080 portunda çalışacağını belirtiyoruz
EXPOSE 8080

# Konteyner çalıştığında uygulamamızı başlat komutu
CMD ["./kan-uygulamasi"]
