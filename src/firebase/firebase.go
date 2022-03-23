package firebase

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"log"
)

var firebaseApp *firebase.App

func SendNotification(topic string, title string, body string, url string) {
	logrus.Info("Send notification")

	ctx := context.Background()
	client, err := firebaseApp.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
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
		log.Fatalln(err)
	}

	logrus.Info(topic, response)
}

func InitFirebaseApp() {
	opt := option.WithCredentialsFile("credentials_firebase.json")
	var err error
	firebaseApp, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		logrus.Errorf("error initializing firebaseApp: %v", err)
	}
}
