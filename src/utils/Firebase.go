package utils

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"os"
)

var firebaseApp *firebase.App

func SendNotification(topic string, title string, body string, url string) error {
	logrus.Info("Send notification")

	ctx := context.Background()
	client, err := firebaseApp.Messaging(ctx)
	if err != nil {
		return fmt.Errorf("error getting Messaging client: %s", err)
	}

	messages := &messaging.Message{
		Notification: &messaging.Notification{
			Title:    title,
			Body:     body,
			ImageURL: url,
		},
		Topic: topic,
	}

	response, err := client.Send(context.Background(), messages)
	if err != nil {
		return err
	}

	logrus.Info(topic, response)
	return nil
}

func InitFirebaseApp() {
	credentials := os.Getenv("FIREBASE_CREDENTIALS")
	opt := option.WithCredentialsJSON([]byte(credentials))

	var err error
	firebaseApp, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		logrus.Errorf("error initializing firebaseApp: %v", err)
	}
}
