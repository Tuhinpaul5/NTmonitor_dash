package mailer

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	from   string
	dialer *gomail.Dialer
}

func New() *Mailer {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))

	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		port,
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
	)
	d.TLSConfig = &tls.Config{ServerName: os.Getenv("SMTP_HOST")}

	return &Mailer{
		from:   os.Getenv("SMTP_FROM"),
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