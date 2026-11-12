package notifications

import (
	"bytes"
	"io"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"testing"
)

func TestTemplateEscapesHTML(t *testing.T) {
	body, err := renderTemplate(adminPasswordResetTemplate, map[string]string{"NewPassword": "<b>&x"})
	if err != nil {
		t.Fatalf("erro ao montar o e-mail: %v", err)
	}

	if !strings.Contains(body, "&lt;b&gt;&amp;x") {
		t.Errorf("a senha não foi escapada no HTML: %s", body)
	}
}

func TestBuildMessage(t *testing.T) {
	from := &mail.Address{Name: "Go API", Address: "no-reply@go.local"}
	htmlBody := "<p>Olá, " + strings.Repeat("ç", 100) + "</p>"

	rawMessage, err := buildMessage(from, "maria@email.com", "Redefinição de senha", htmlBody)
	if err != nil {
		t.Fatalf("erro ao montar a mensagem: %v", err)
	}

	message, err := mail.ReadMessage(bytes.NewReader(rawMessage))
	if err != nil {
		t.Fatalf("a mensagem gerada não é um e-mail válido: %v", err)
	}

	subject, err := new(mime.WordDecoder).DecodeHeader(message.Header.Get("Subject"))
	if err != nil || subject != "Redefinição de senha" {
		t.Errorf("assunto = %q (erro %v), esperado %q", subject, err, "Redefinição de senha")
	}

	if message.Header.Get("To") != "maria@email.com" {
		t.Errorf("destinatário = %q", message.Header.Get("To"))
	}

	decodedBody, err := io.ReadAll(quotedprintable.NewReader(message.Body))
	if err != nil {
		t.Fatalf("erro ao decodificar o corpo: %v", err)
	}

	if string(decodedBody) != htmlBody {
		t.Errorf("corpo decodificado diferente do original")
	}

	for _, line := range strings.Split(string(rawMessage), "\r\n") {
		if len(line) > 78 {
			t.Errorf("linha com %d caracteres, o limite de e-mail é 78: %q", len(line), line)
		}
	}
}
