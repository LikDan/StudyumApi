package mail

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
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
	ForceSend(to, subject, body string)

	SendFile(to, subject, filename string, data Data) error
	ForceSendFile(to, subject, filename string, data Data)
}

type mail struct {
	Service *gmail.Service

	Mode         Mode
	templatesDir string
}

func NewMail(ctx context.Context, mode Mode, id, secret, access, refresh, templatesDir string) Mail {
	mail := mail{Mode: mode, templatesDir: templatesDir}

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

func (m *mail) send(message gmail.Message) error {
	_, err := m.Service.Users.Messages.Send("me", &message).Do()
	return err
}

func (m *mail) Send(to, subject, body string) error {
	if m.Mode == DebugMode {
		logrus.Debugln("----------Email----------")
		logrus.Debugln()
		logrus.Debugln("Sending email to " + to)
		logrus.Debugln("Subject: " + subject)
		logrus.Debugln("Body:")
		logrus.Debugln(body)
		logrus.Debugln()

		return nil
	}

	message := m.buildMessage(to, subject, body)
	return m.send(message)
}

func (m *mail) ForceSend(to, subject, body string) {
	if err := m.Send(to, subject, body); err != nil {
		panic("error sending email " + err.Error())
	}
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

func (m *mail) ForceSendFile(to, subject, filename string, data Data) {
	if err := m.SendFile(to, subject, filename, data); err != nil {
		panic("error sending file email " + err.Error())
	}
}
