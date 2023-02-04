package main

import (
	"context"
	"fmt"
	"os"
	"studyum/pkg/mail"
)

func main() {
	id := os.Getenv("GMAIL_CLIENT_ID")
	secret := os.Getenv("GMAIL_CLIENT_SECRET")
	access := os.Getenv("GMAIL_ACCESS_TOKEN")
	refresh := os.Getenv("GMAIL_REFRESH_TOKEN")
	m := mail.NewMail(context.Background(), id, secret, access, refresh, "email-templates")
	if err := m.SendFile("likdan.official@gmail.com", "Confirmation code", "code.html", mail.Data{"code": "000-000", "name": "LikDan", "expire": "01.01.2000 10:10"}); err != nil {
		fmt.Println(err)
		return
	}
}
