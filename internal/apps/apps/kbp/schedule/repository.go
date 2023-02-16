package schedule

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	html "github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Repository interface {
	AddLesson(ctx context.Context, token string, lesson Lesson) (string, error)
	UpdateLesson(ctx context.Context, token string, lesson Lesson) (string, error)
	DeleteLesson(ctx context.Context, token string, lesson Lesson) (string, error)
}

type repository struct {
}

func NewRepository() Repository {
	return &repository{}
}

func (r *repository) send(ctx context.Context, token string, lesson Lesson) (string, error) {
	requestBytes, _ := json.Marshal(lesson)
	requestData := string(requestBytes)
	requestData = strings.ReplaceAll(requestData, ":", "=")
	requestData = strings.ReplaceAll(requestData, ",", "&")
	requestData = strings.ReplaceAll(requestData, "\"", "")
	requestData = requestData[1 : len(requestData)-1]

	if os.Getenv("KBP_MODE") != "release" {
		logrus.Info("Lesson kbp " + requestData)
		return "", errors.New("app running in debug mode")
	}

	request, _ := http.NewRequestWithContext(ctx, "POST", "https://ej.kbp.by/ajax.php", bytes.NewBuffer([]byte(requestData)))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(requestData)))
	request.Header.Add("Cookie", token)
	request.Header.Add("Host", "kbp.by")

	response, _ := http.DefaultClient.Do(request)

	doc, _ := html.NewDocumentFromReader(response.Body)
	id, ok := doc.Find("#dateOfMonth").First().Find("div[on-date=\"" + lesson.Date + "\"]").Last().Attr("pair-id")
	if !ok {
		return "", errors.New("no date found")
	}

	return id, nil
}

func (r *repository) AddLesson(ctx context.Context, token string, lesson Lesson) (string, error) {
	lesson.Action = "add_date"
	return r.send(ctx, token, lesson)
}

func (r *repository) UpdateLesson(ctx context.Context, token string, lesson Lesson) (string, error) {
	lesson.Action = "edit_date"
	return r.send(ctx, token, lesson)
}

func (r *repository) DeleteLesson(context.Context, string, Lesson) (string, error) {
	return "", nil
}
