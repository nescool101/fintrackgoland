package service

import (
	"fmt"
	"io"
	"log"
	"time"

	"gopkg.in/gomail.v2"
)

// EmailService manejo de envío de correos electrónicos
type EmailService struct {
	host     string
	port     int
	username string
	password string
}

// NewEmailService crea una nueva instancia del servicio de email
func NewEmailService(host string, port int, username, password string) *EmailService {
	return &EmailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

// SendExcelReport envía un reporte Excel por correo electrónico
func (es *EmailService) SendExcelReport(recipient, subject, body string, excelData []byte, filename string) error {
	// Crear mensaje
	m := gomail.NewMessage()

	// Configurar remitente y destinatario
	m.SetHeader("From", es.username)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)

	// Configurar cuerpo del mensaje
	m.SetBody("text/html", es.buildHTMLBody(body))

	// Adjuntar archivo Excel
	m.Attach(filename, gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(excelData)
		return err
	}))

	// Configurar dialer SMTP
	d := gomail.NewDialer(es.host, es.port, es.username, es.password)

	// Enviar el correo
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error enviando correo: %v", err)
	}

	log.Printf("📧 Correo enviado exitosamente a: %s", recipient)
	return nil
}

// buildHTMLBody construye el cuerpo HTML del correo
func (es *EmailService) buildHTMLBody(message string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: #4CAF50;
            color: white;
            padding: 20px;
            text-align: center;
            border-radius: 5px;
        }
        .content {
            padding: 20px;
            background-color: #f9f9f9;
            border-radius: 5px;
            margin: 20px 0;
        }
        .footer {
            text-align: center;
            font-size: 12px;
            color: #666;
            margin-top: 20px;
        }
        .highlight {
            background-color: #ffeb3b;
            padding: 2px 4px;
            border-radius: 3px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>📊 Reporte Financiero FinTrack</h1>
    </div>
    
    <div class="content">
        <p>Estimado usuario,</p>
        
        <p>%s</p>
        
        <p>En el archivo Excel adjunto encontrará:</p>
        <ul>
            <li><strong>Hoja "Reporte_Financiero":</strong> Datos detallados de stocks e índices</li>
            <li><strong>Hoja "Resumen":</strong> Estadísticas generales del reporte</li>
        </ul>
        
        <p>Los datos incluyen información de apertura, cierre, máximos, mínimos y volumen para cada símbolo.</p>
        
        <p class="highlight">📈 Símbolos incluidos: SPX, NDX, DJI, NYA, ES_F, NQ_F</p>
    </div>
    
    <div class="footer">
        <p>Generado automáticamente por FinTrack GoLand</p>
        <p>Fecha: %s</p>
    </div>
</body>
</html>
`, message, time.Now().Format("2006-01-02 15:04:05"))
}

// SendSimpleEmail envía un correo simple sin adjuntos
func (es *EmailService) SendSimpleEmail(recipient, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", es.username)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(es.host, es.port, es.username, es.password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error enviando correo simple: %v", err)
	}

	log.Printf("📧 Correo simple enviado a: %s", recipient)
	return nil
}
