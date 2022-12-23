package controller

import (
	"bytes"
	"context"
	htmlParser "github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"strconv"
	"strings"
	"studyum/internal/parser/apps/kbp/entities"
	"studyum/internal/parser/apps/kbp/repository"
	entities2 "studyum/internal/parser/entities"
)

type Controller struct {
	repository *repository.Repository
	http       *repository.JournalHTTP
}

func NewController(r *repository.Repository) *Controller {
	return &Controller{repository: r, http: repository.NewJournalHTTP()}
}

func (c *Controller) loginJournalTeacher(user entities.JournalUser) string {
	request, _ := http.NewRequest("GET", "https://nehai.by/ej/templates/login_parent.php", nil)
	response, _ := http.DefaultClient.Do(request)
	document, _ := htmlParser.NewDocumentFromReader(response.Body)
	sCode, _ := document.Find("#S_Code").Attr("value")

	cookie := response.Cookies()[0]

	requestString := "action=login_teather&login=" + user.Login + "&password=" + user.Password + "&S_Code=" + sCode
	responseBody := bytes.NewBuffer([]byte(requestString))
	request, _ = http.NewRequest("POST", "https://nehai.by/ej/ajax.php", responseBody)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(requestString)))
	request.Header.Add("Host", "kbp.by")
	cookieStr := cookie.Name + "=" + cookie.Value
	request.Header.Add("Cookie", cookieStr)
	response, _ = http.DefaultClient.Do(request)

	body, _ := io.ReadAll(response.Body)
	if string(body) != "good" {
		return ""
	}

	return cookieStr
}

func (c *Controller) AddMark(ctx context.Context, mark entities2.Mark, lesson entities2.Lesson, student entities2.User) error {
	err, user := c.repository.GetJournalUserByID(ctx, lesson.Teacher)
	if err != nil {
		return err
	}

	cookie := c.loginJournalTeacher(user)

	options := c.http.GetJournalTeacherOptions(ctx, cookie)
	groupID := options.Find(".group").FilterFunction(func(i int, selection *htmlParser.Selection) bool {
		a := selection.Text()[:2]
		b := strings.ToLower(lesson.Group[len(lesson.Group)-2:])
		c := selection.Text()[len(selection.Text())-(len(lesson.Group)-2):]
		d := lesson.Group[:2]
		letter := a == b && c == d
		return letter
	}).AttrOr("groupid", "")

	subjectID := options.Find(".subject").FilterFunction(func(i int, selection *htmlParser.Selection) bool {
		return selection.AttrOr("groupid", "") == groupID && selection.Text() == lesson.Subject
	}).AttrOr("subjectid", "")

	journal := c.http.GetJournalTeacher(ctx, cookie, subjectID, groupID)
	studentID := journal.Find(".leftTable").Find(".pupilName").FilterFunction(func(i int, selection *htmlParser.Selection) bool {
		return strings.TrimSpace(selection.Text()) == student.Name
	}).First().Parent().Parent().AttrOr("class", "row")[3:]

	pairID := journal.Find(".rightTable").Find("#dateOfMonth").Children().FilterFunction(func(i int, selection *htmlParser.Selection) bool {
		return selection.Children().AttrOr("on-date", "") == lesson.StartDate.Format("2006-1-2")
	}).First().Children().AttrOr("pair-id", "")

	err = c.http.AddMark(ctx, cookie, studentID, pairID, mark.Mark)
	if err != nil {
		return err
	}

	return nil
}
