package services

import (
	"NTMonitor/config"
	"crypto/tls"
	"fmt"
	"strconv"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	from   string
	dialer *gomail.Dialer
}

func New(cfg *config.Config) *Mailer {
	port, _ := strconv.Atoi(cfg.SMTP_PORT)

	d := gomail.NewDialer(
		cfg.SMTP_HOST,
		port,
		cfg.SMTP_USERNAME,
		cfg.SMTP_PASSWORD,
	)
	d.TLSConfig = &tls.Config{ServerName: cfg.SMTP_HOST}

	return &Mailer{
		from:   cfg.SMTP_FROM,
		dialer: d,
	}
}

func (m *Mailer) Send(to, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}