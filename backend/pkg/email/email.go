package email

import (
	"gopkg.in/gomail.v2"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type EmailService struct {
	config SMTPConfig
}

func NewEmailService(cfg SMTPConfig) *EmailService {
	return &EmailService{config: cfg}
}

func (s *EmailService) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)
	return d.DialAndSend(m)
}
