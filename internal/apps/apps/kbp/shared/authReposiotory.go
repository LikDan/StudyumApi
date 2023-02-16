package shared

import (
	"bytes"
	"context"
	html "github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"strconv"
)

type AuthRepository interface {
	Auth(ctx context.Context) string
}

type authRepository struct {
	login, password string
}

func NewAuthRepository(login string, password string) AuthRepository {
	return &authRepository{login: login, password: password}
}

func (r *authRepository) Auth(ctx context.Context) string {
	request, _ := http.NewRequestWithContext(ctx, "GET", "https://ej.kbp.by/templates/login_parent.php", nil)
	response, _ := http.DefaultClient.Do(request)
	document, _ := html.NewDocumentFromReader(response.Body)
	sCode, _ := document.Find("#S_Code").Attr("value")

	cookie := response.Cookies()[0]

	requestData := "action=login_teather&login=" + r.login + "&password=" + r.password + "&S_Code=" + sCode

	request, _ = http.NewRequestWithContext(ctx, "POST", "https://ej.kbp.by/ajax.php", bytes.NewBuffer([]byte(requestData)))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(requestData)))
	request.Header.Add("Host", "kbp.by")

	token := cookie.Name + "=" + cookie.Value
	request.Header.Add("Cookie", token)
	response, _ = http.DefaultClient.Do(request)

	body, _ := io.ReadAll(response.Body)
	if string(body) != "good" {
		return ""
	}

	return token
}
