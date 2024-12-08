package utils

import (
	"gopkg.in/gomail.v2"
	"io"
	"log"
)

func SendEmail(host string, port int, user, pass, to, subject, body string, attachment []byte, filename string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", user)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	if attachment != nil && filename != "" {
		m.Attach(filename, gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(attachment)
			return err
		}))
	}

	d := gomail.NewDialer(host, port, user, pass)

	// Optional: Customize TLS settings if necessary
	// d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	log.Println("Email sent successfully")
	return nil
}
