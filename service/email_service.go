package service

import (
	"fmt"
	"io"
	"log"
	"strings"
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
func (es *EmailService) SendExcelReport(recipient, subject, body string, excelData []byte, filename string, symbolCount int) error {
	// Lista de destinatarios (siempre incluir ambos emails)
	recipients := []string{
		"nescool101@gmail.com",
		"paulocesarcelis@gmail.com",
	}

	// Si se especifica un destinatario diferente, agregarlo también
	if recipient != "" && recipient != "nescool101@gmail.com" && recipient != "paulocesarcelis@gmail.com" {
		// Verificar si el recipient contiene múltiples emails separados por coma
		if strings.Contains(recipient, ",") {
			additionalRecipients := strings.Split(recipient, ",")
			for _, addr := range additionalRecipients {
				addr = strings.TrimSpace(addr)
				if addr != "" && addr != "nescool101@gmail.com" && addr != "paulocesarcelis@gmail.com" {
					recipients = append(recipients, addr)
				}
			}
		} else {
			recipients = append(recipients, recipient)
		}
	}

	// Configurar dialer SMTP
	d := gomail.NewDialer(es.host, es.port, es.username, es.password)

	var sentTo []string
	var errors []string

	// Enviar email individual a cada destinatario
	for _, recipientAddr := range recipients {
		// Crear mensaje individual
		m := gomail.NewMessage()
		m.SetHeader("From", es.username)
		m.SetHeader("To", recipientAddr)
		m.SetHeader("Subject", subject)

		// Configurar cuerpo del mensaje
		m.SetBody("text/html", es.buildHTMLBody(body, symbolCount))

		// Adjuntar archivo Excel
		m.Attach(filename, gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(excelData)
			return err
		}))

		// Enviar el correo individual
		if err := d.DialAndSend(m); err != nil {
			errorMsg := fmt.Sprintf("Error enviando a %s: %v", recipientAddr, err)
			errors = append(errors, errorMsg)
			log.Printf("❌ %s", errorMsg)
		} else {
			sentTo = append(sentTo, recipientAddr)
			log.Printf("✅ Correo enviado exitosamente a: %s", recipientAddr)
		}
	}

	// Verificar si se envió al menos un correo
	if len(sentTo) == 0 {
		return fmt.Errorf("no se pudo enviar el correo a ningún destinatario: %s", strings.Join(errors, "; "))
	}

	// Si hubo algunos errores pero al menos uno se envió exitosamente
	if len(errors) > 0 {
		log.Printf("⚠️ Algunos correos fallaron: %s", strings.Join(errors, "; "))
	}

	log.Printf("📧 Correo enviado exitosamente a: %v (%d de %d destinatarios)", sentTo, len(sentTo), len(recipients))
	return nil
}

// buildHTMLBody construye el cuerpo HTML del correo
func (es *EmailService) buildHTMLBody(message string, symbolCount int) string {
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
        .warning {
            background-color: #fff3cd;
            border: 1px solid #ffeaa7;
            color: #856404;
            padding: 10px;
            border-radius: 5px;
            margin: 10px 0;
        }
        .info {
            background-color: #d1ecf1;
            border: 1px solid #bee5eb;
            color: #0c5460;
            padding: 10px;
            border-radius: 5px;
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>📊 Reporte Financiero FinTrack</h1>
    </div>
    
    <div class="content">
        <p>Estimado usuario,</p>
        
        <div>%s</div>
        
        <p>En el archivo Excel adjunto encontrará:</p>
        <ul>
            <li><strong>Hoja "Reporte_Financiero":</strong> Todos los datos combinados</li>
            <li><strong>Hoja "Stocks":</strong> Datos específicos de stocks</li>
            <li><strong>Hoja "Indices":</strong> Datos específicos de índices</li>
            <li><strong>Hoja "Resumen":</strong> Estadísticas generales del reporte</li>
        </ul>
        
        <p>Los datos incluyen información de apertura, cierre, máximos, mínimos y volumen para cada símbolo.</p>
        
        <p class="highlight">📈 Símbolos procesados: %d símbolos totales</p>
    </div>
    
    <div class="footer">
        <p>Generado automáticamente por FinTrack GoLand</p>
        <p>Fecha: %s</p>
    </div>
</body>
</html>
`, message, symbolCount, time.Now().Format("2006-01-02 15:04:05"))
}

// SendSimpleEmail envía un correo simple sin adjuntos
func (es *EmailService) SendSimpleEmail(recipient, subject, body string) error {
	// Lista de destinatarios (siempre incluir ambos emails)
	recipients := []string{
		"nescool101@gmail.com",
		"paulocesarcelis@gmail.com",
	}

	// Si se especifica un destinatario diferente, agregarlo también
	if recipient != "" && recipient != "nescool101@gmail.com" && recipient != "paulocesarcelis@gmail.com" {
		// Verificar si el recipient contiene múltiples emails separados por coma
		if strings.Contains(recipient, ",") {
			additionalRecipients := strings.Split(recipient, ",")
			for _, addr := range additionalRecipients {
				addr = strings.TrimSpace(addr)
				if addr != "" && addr != "nescool101@gmail.com" && addr != "paulocesarcelis@gmail.com" {
					recipients = append(recipients, addr)
				}
			}
		} else {
			recipients = append(recipients, recipient)
		}
	}

	d := gomail.NewDialer(es.host, es.port, es.username, es.password)

	var sentTo []string
	var errors []string

	// Enviar email individual a cada destinatario
	for _, recipientAddr := range recipients {
		m := gomail.NewMessage()
		m.SetHeader("From", es.username)
		m.SetHeader("To", recipientAddr)
		m.SetHeader("Subject", subject)
		m.SetBody("text/plain", body)

		if err := d.DialAndSend(m); err != nil {
			errorMsg := fmt.Sprintf("Error enviando a %s: %v", recipientAddr, err)
			errors = append(errors, errorMsg)
			log.Printf("❌ %s", errorMsg)
		} else {
			sentTo = append(sentTo, recipientAddr)
			log.Printf("✅ Correo simple enviado a: %s", recipientAddr)
		}
	}

	// Verificar si se envió al menos un correo
	if len(sentTo) == 0 {
		return fmt.Errorf("no se pudo enviar el correo a ningún destinatario: %s", strings.Join(errors, "; "))
	}

	// Si hubo algunos errores pero al menos uno se envió exitosamente
	if len(errors) > 0 {
		log.Printf("⚠️ Algunos correos simples fallaron: %s", strings.Join(errors, "; "))
	}

	log.Printf("📧 Correo simple enviado a: %v (%d de %d destinatarios)", sentTo, len(sentTo), len(recipients))
	return nil
}
