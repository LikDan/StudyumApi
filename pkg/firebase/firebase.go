package firebase

import (
	"context"
	fb "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

type Firebase interface {
	SendNotification(topic string, title string, body string, url string) (string, error)
}

type firebase struct {
	app *fb.App
}

func NewFirebase(credentials []byte) Firebase {
	opt := option.WithCredentialsJSON(credentials)

	app, err := fb.NewApp(context.Background(), nil, opt)
	if err != nil {
		logrus.Errorf("error initializing firebaseApp: %v", err)
		return nil
	}

	return &firebase{app: app}
}

func (f *firebase) SendNotification(topic string, title string, body string, url string) (string, error) {
	ctx := context.Background()
	client, err := f.app.Messaging(ctx)
	if err != nil {
		return "", err
	}

	messages := &messaging.Message{
		Notification: &messaging.Notification{
			Title:    title,
			Body:     body,
			ImageURL: url,
		},
		Topic: topic,
	}

	return client.Send(ctx, messages)
}
