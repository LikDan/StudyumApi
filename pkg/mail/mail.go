package mail

import (
	"context"
	"encoding/base64"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"os"
	"strings"
	"time"
)

const rawMessage = `To: %s
Subject: %s
MIME-version: 1.0;
Content-Type: text/html; charset="UTF-8";


%s`

type Mail interface {
	Send(to, subject, body string) error
	SendFile(to, subject, filename string, data Data) error
}

type mail struct {
	Service *gmail.Service

	templatesDir string
}

func NewMail(ctx context.Context, id, secret, access, refresh, templatesDir string) Mail {
	mail := mail{templatesDir: templatesDir}

	service, err := mail.init(ctx, id, secret, access, refresh)
	if err != nil {
		return nil
	}

	mail.Service = service
	return &mail
}

func (m *mail) init(ctx context.Context, id, secret, access, refresh string) (*gmail.Service, error) {
	config := oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		Endpoint:     google.Endpoint,
	}

	token := oauth2.Token{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		Expiry:       time.Now(),
	}

	var tokenSource = config.TokenSource(ctx, &token)
	return gmail.NewService(ctx, option.WithTokenSource(tokenSource))
}

func (m *mail) buildMessage(to, subject, body string) gmail.Message {
	message := fmt.Sprintf(rawMessage, to, subject, body)
	return gmail.Message{Raw: base64.URLEncoding.EncodeToString([]byte(message))}
}

func (m *mail) Send(to, subject, body string) error {
	message := m.buildMessage(to, subject, body)

	_, err := m.Service.Users.Messages.Send("me", &message).Do()
	return err
}

func (m *mail) proceedFile(filename string, data map[string]string) (string, error) {
	bytes, err := os.ReadFile(m.templatesDir + "/" + filename)
	text := string(bytes)

	for key, val := range data {
		text = strings.ReplaceAll(text, "{"+key+"}", val)
	}
	text = strings.ReplaceAll(text, "\\}", "}")

	return text, err
}

func (m *mail) SendFile(to, subject, filename string, data Data) error {
	body, err := m.proceedFile(filename, data)
	if err != nil {
		return err
	}

	return m.Send(to, subject, body)
}
