package main

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"fmt"
	"google.golang.org/api/option"
	"log"
)

var firebaseApp *firebase.App

func sendNotification(topic string, title string, body string, url string) {
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

	fmt.Println(topic, response)
}

func initFirebaseApp() {
	opt := option.WithCredentialsFile("credentials_firebase.json")
	var err error
	firebaseApp, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing firebaseApp: %v\n", err)
	}
}
