package messaging

import (
	"fmt"
	"net/smtp"

	"github.com/chegde20121/Tigerhall-Kittens/pkg/config"
	log "github.com/sirupsen/logrus"
)

type EmailHandler struct {
	logger *log.Logger
}

type EmailTemplateData struct {
	TigerName     string
	TigerLocation string
	SightingTime  string
	Organization  string
	ContactInfo   string
}

func NewEmailHandler(logger *log.Logger) *EmailHandler {
	return &EmailHandler{logger: logger}
}

type EmailStatus struct {
	Err   error
	Email string
}

func (em *EmailHandler) SendEmailNotification(emails []string, subject string, body string) (err error) {
	em.logger.Info("Sending Email Notification")
	senderEmail := config.GetEnvVar("SENDER_EMAIL")
	senderPassword := config.GetEnvVar("SENDER_EMAIL_PASSWORD")
	smtpHost := config.GetEnvVar("SMTP_HOST")
	smtpPort := config.GetEnvVar("SMTP_PORT")
	emailChan := make(chan EmailStatus)
	message := "Subject: " + subject + "\r\n" + "\r\n" + body
	errorMessage := []EmailStatus{}
	for _, email := range emails {
		auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpHost)
		to := email
		go func(from, to, message string) {
			err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
			emailChan <- EmailStatus{
				Err:   err,
				Email: to,
			}
		}(senderEmail, to, message)
	}
	for range emails {
		status := <-emailChan
		if status.Err != nil {
			errorMessage = append(errorMessage, status)
		}
	}
	if len(errorMessage) > 0 {
		err = fmt.Errorf("failed to send email notification:[%v]", errorMessage)
	}
	return
}
