package marks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Repository interface {
	AddMark(ctx context.Context, token string, mark Mark) (string, error)
	UpdateMark(ctx context.Context, token string, mark Mark) (string, error)
	DeleteMark(ctx context.Context, token string, mark Mark) (string, error)
}

type repository struct {
}

func (r *repository) send(ctx context.Context, token string, mark Mark) (string, error) {
	requestBytes, _ := json.Marshal(mark)
	requestData := string(requestBytes)
	requestData = strings.ReplaceAll(requestData, ":", "=")
	requestData = strings.ReplaceAll(requestData, ",", "&")
	requestData = strings.ReplaceAll(requestData, "\"", "")
	requestData = requestData[1 : len(requestData)-1]

	if os.Getenv("KBP_MODE") != "release" {
		logrus.Info("Marks kbp " + requestData)
		return "", errors.New("app running in debug mode")
	}

	request, _ := http.NewRequestWithContext(ctx, "POST", "https://ej.kbp.by/ajax.php", bytes.NewBuffer([]byte(requestData)))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(requestData)))
	request.Header.Add("Cookie", token)
	request.Header.Add("Host", "kbp.by")

	response, _ := http.DefaultClient.Do(request)

	bodyBytes, _ := io.ReadAll(response.Body)
	body := string(bodyBytes)
	body = body[strings.LastIndex(body, "\n")+1:]
	return body, nil
}

func (r *repository) AddMark(ctx context.Context, token string, mark Mark) (string, error) {
	mark.Action = "set_mark"
	mark.MarkID = "0"
	return r.send(ctx, token, mark)
}

func (r *repository) UpdateMark(ctx context.Context, token string, mark Mark) (string, error) {
	mark.Action = "set_mark"
	return r.send(ctx, token, mark)
}

func (r *repository) DeleteMark(ctx context.Context, token string, mark Mark) (string, error) {
	mark.Action = "set_mark"
	mark.Value = "X"
	return r.send(ctx, token, mark)
}
