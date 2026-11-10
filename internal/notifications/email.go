package notifications

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/PietroTelino/go-api-template/internal/config"
)

var passwordResetTemplate = template.Must(template.New("passwordReset").Parse(`
<p>Você solicitou a redefinição de sua senha.</p>
<p>Clique no link abaixo para criar uma nova senha:</p>
<p><a href="{{.ResetLink}}">Redefinir senha</a></p>
<p>Se você não solicitou isso, ignore este e-mail.</p>
`))

var adminPasswordResetTemplate = template.Must(template.New("adminPasswordReset").Parse(`
<p>Sua senha foi resetada por um administrador.</p>
<p>Sua nova senha temporária é:</p>
<p style="font-size: 18px; font-weight: bold; background-color: #f0f0f0; padding: 10px; border-radius: 5px; display: inline-block;">{{.NewPassword}}</p>
<p>Por motivos de segurança, recomendamos que você altere esta senha assim que fizer login.</p>
<p>Se você não solicitou essa alteração, entre em contato com o suporte imediatamente.</p>
`))

type EmailService struct {
	enabled  bool
	host     string
	port     int
	secure   bool
	user     string
	password string
	from     string
}

func NewEmailService(appConfig *config.Config) *EmailService {
	if !appConfig.EmailEnabled {
		log.Println("[EmailService] Envio de e-mail desabilitado")
	}

	return &EmailService{
		enabled:  appConfig.EmailEnabled,
		host:     appConfig.EmailHost,
		port:     appConfig.EmailPort,
		secure:   appConfig.EmailSecure,
		user:     appConfig.EmailUser,
		password: appConfig.EmailPassword,
		from:     appConfig.EmailFrom,
	}
}

func (emailService *EmailService) SendPasswordReset(to string, resetLink string) error {
	if !emailService.enabled {
		log.Printf("[EmailService] (DEV MODE) Para: %s | Link de redefinição: %s", to, resetLink)
		return nil
	}

	body, err := renderTemplate(passwordResetTemplate, map[string]string{"ResetLink": resetLink})
	if err != nil {
		return err
	}

	return emailService.send(to, "Redefinição de senha", body)
}

func (emailService *EmailService) SendAdminPasswordReset(to string, newPassword string) error {
	if !emailService.enabled {
		log.Printf("[EmailService] (DEV MODE) Para: %s | Nova senha: %s", to, newPassword)
		return nil
	}

	body, err := renderTemplate(adminPasswordResetTemplate, map[string]string{"NewPassword": newPassword})
	if err != nil {
		return err
	}

	return emailService.send(to, "Sua senha foi resetada", body)
}

func renderTemplate(htmlTemplate *template.Template, data any) (string, error) {
	var buffer bytes.Buffer

	if err := htmlTemplate.Execute(&buffer, data); err != nil {
		return "", fmt.Errorf("erro ao montar o e-mail: %w", err)
	}

	return buffer.String(), nil
}

func buildMessage(from *mail.Address, to string, subject string, htmlBody string) ([]byte, error) {
	var buffer bytes.Buffer

	headers := []string{
		"From: " + from.String(),
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("UTF-8", subject),
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="UTF-8"`,
		"Content-Transfer-Encoding: quoted-printable",
	}

	buffer.WriteString(strings.Join(headers, "\r\n") + "\r\n\r\n")

	bodyWriter := quotedprintable.NewWriter(&buffer)
	if _, err := bodyWriter.Write([]byte(htmlBody)); err != nil {
		return nil, err
	}

	if err := bodyWriter.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (emailService *EmailService) send(to string, subject string, htmlBody string) error {
	fromAddress, err := mail.ParseAddress(emailService.from)
	if err != nil {
		return fmt.Errorf("EMAIL_FROM inválido: %w", err)
	}

	message, err := buildMessage(fromAddress, to, subject, htmlBody)
	if err != nil {
		return fmt.Errorf("erro ao montar o e-mail: %w", err)
	}

	address := net.JoinHostPort(emailService.host, strconv.Itoa(emailService.port))
	tlsConfig := &tls.Config{ServerName: emailService.host}
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	var connection net.Conn
	if emailService.secure {
		connection, err = tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	} else {
		connection, err = dialer.Dial("tcp", address)
	}

	if err != nil {
		return fmt.Errorf("erro ao conectar ao servidor SMTP: %w", err)
	}

	if err := connection.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		connection.Close()
		return err
	}

	client, err := smtp.NewClient(connection, emailService.host)
	if err != nil {
		connection.Close()
		return fmt.Errorf("erro ao iniciar a conversa SMTP: %w", err)
	}
	defer client.Close()

	if !emailService.secure {
		if supportsStartTLS, _ := client.Extension("STARTTLS"); supportsStartTLS {
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("erro ao iniciar o STARTTLS: %w", err)
			}
		}
	}

	if emailService.user != "" {
		auth := smtp.PlainAuth("", emailService.user, emailService.password, emailService.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("erro ao autenticar no servidor SMTP: %w", err)
		}
	}

	if err := client.Mail(fromAddress.Address); err != nil {
		return fmt.Errorf("erro no remetente: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("erro no destinatário: %w", err)
	}

	dataWriter, err := client.Data()
	if err != nil {
		return fmt.Errorf("erro ao enviar o conteúdo: %w", err)
	}

	if _, err := dataWriter.Write(message); err != nil {
		return fmt.Errorf("erro ao enviar o conteúdo: %w", err)
	}

	if err := dataWriter.Close(); err != nil {
		return fmt.Errorf("erro ao finalizar o envio: %w", err)
	}

	return client.Quit()
}
