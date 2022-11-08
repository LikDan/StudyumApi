package repository

import (
	"bytes"
	"context"
	htmlParser "github.com/PuerkitoBio/goquery"
	"net/http"
)

type JournalHTTP struct {
}

func NewJournalHTTP() *JournalHTTP {
	return &JournalHTTP{}
}

func (r *JournalHTTP) AddMark(ctx context.Context, cookie string, studentID, pairID, mark string) error {
	requestString := "action=set_mark&student_id=" + studentID + "&pair_id=" + pairID + "&mark_id=0&value=" + mark
	responseBody := bytes.NewBuffer([]byte(requestString))
	request, _ := http.NewRequestWithContext(ctx, "POST", "https://nehai.by/ej/ajax.php", responseBody)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Cookie", cookie)

	return nil
	_, err := http.DefaultClient.Do(request)
	return err
}

func (r *JournalHTTP) GetJournalTeacherOptions(ctx context.Context, cookie string) *htmlParser.Document {
	request, _ := http.NewRequestWithContext(ctx, "GET", "https://nehai.by/ej/templates/teacher_journal.php", nil)
	request.Header.Add("Cookie", cookie)
	request.Header.Add("Host", "nehai.by")
	response, _ := http.DefaultClient.Do(request)

	document, _ := htmlParser.NewDocumentFromReader(response.Body)
	return document
}

func (r *JournalHTTP) GetJournalTeacher(ctx context.Context, cookie, subjectID, groupID string) *htmlParser.Document {
	requestString := "action=show_table&subject_id=" + subjectID + "&group_id=" + groupID
	responseBody := bytes.NewBuffer([]byte(requestString))
	request, _ := http.NewRequestWithContext(ctx, "POST", "https://nehai.by/ej/ajax.php", responseBody)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Cookie", cookie)
	response, _ := http.DefaultClient.Do(request)

	document, _ := htmlParser.NewDocumentFromReader(response.Body)
	return document
}
