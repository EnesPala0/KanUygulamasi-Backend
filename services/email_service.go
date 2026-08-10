package services

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

// SendEmaille resend API'sini kullıcaz ve e-posta göndercez
func SendEmail(to string, subject string, htmlContent string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY environment variable is not set")
	}

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "KanBağı App <onboarding@resend.dev>",
		To:      []string{to}, // Şimdilik sadece kendi mailine gidecek
		Subject: subject,
		Html:    htmlContent,
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("mail gönderilemedi (Resend API Hatası): %v", err)
	}

	fmt.Printf("Mail başarıyla fırlatıldı! Resend ID: %s\n", sent.Id)
	return nil

}
