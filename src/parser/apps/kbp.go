package apps

import (
	"bytes"
	htmlParser "github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	h "studyum/src/api"
	"studyum/src/db"
	"studyum/src/models"
	"time"
)

type KbpParser struct {
	States     []*models.ScheduleStateInfo
	TempStates []*models.ScheduleStateInfo

	WeekdaysShift []h.Shift
	WeekendsShift []h.Shift
}

var KbpApp = KbpParser{}

func (p *KbpParser) GetName() string              { return "kbp" }
func (p *KbpParser) GetStudyPlaceId() int         { return 0 }
func (p *KbpParser) GetUpdateCronPattern() string { return "@every 30m" }

func (p *KbpParser) ScheduleUpdate(type_ *models.ScheduleTypeInfo) []*models.Lesson {
	response, err := http.Get("http://kbp.by/rasp/timetable/view_beta_kbp/" + type_.Url)
	if models.BindError(err, 418, h.UNDEFINED).Check() {
		return nil
	}

	if response.StatusCode != 200 {
		models.BindErrorStr("Could not connect to host http://kbp", response.StatusCode, h.UNDEFINED).Check()
		return nil
	}

	document, err := htmlParser.NewDocumentFromReader(response.Body)
	h.CheckError(err, h.WARNING)

	time_ := time.Now().AddDate(0, 0, -int(time.Now().Weekday())).Round(0)

	weeks := document.Find(".find_block").Children().Last().Children()
	if weeks == nil {
		return nil
	}

	var lessons []*models.Lesson

	var states []*models.ScheduleStateInfo

	weeks.Each(func(tableIndex int, table *htmlParser.Selection) {
		weekDate := time.Now().AddDate(0, 0, tableIndex*7)
		if weekDate.Weekday() == time.Sunday {
			weekDate.AddDate(0, 0, 1)
		}
		_, weekIndex := weekDate.ISOWeek()
		weekIndex %= 2

		table.Find("tr").Each(func(rowIndex int, row *htmlParser.Selection) {
			if rowIndex == 1 {
				row.Find("th").Each(func(columnIndex int, selection *htmlParser.Selection) {
					if columnIndex == 0 || columnIndex > 6 {
						return
					}

					stateInfo := models.ScheduleStateInfo{
						WeekIndex: weekIndex,
						DayIndex:  columnIndex - 1,
					}

					state := strings.TrimSpace(selection.Text())
					if state == "" {
						stateInfo.State = models.NotUpdated
					} else {
						stateInfo.State = models.Updated
					}

					states = append(states, &stateInfo)
				})
				return
			}

			rowIndex -= 2
			if rowIndex < 0 {
				return
			}

			row.Find("td").Each(func(columnIndex int, column *htmlParser.Selection) {
				columnIndex -= 1
				if columnIndex < 0 || columnIndex > 5 {
					return
				}

				time_ = time_.AddDate(0, 0, 1)
				column.Find(".pair").Each(func(_ int, div *htmlParser.Selection) {
					var type_ string

					if div.HasClass("added") {
						type_ = "ADDED"
					} else if div.HasClass("removed") && models.GetScheduleStateInfoByIndexes(weekIndex, columnIndex, states).State == models.Updated {
						type_ = "REMOVED"
					} else {
						type_ = "STAY"
					}

					div.Find(".teacher").Each(func(_ int, teacherDiv *htmlParser.Selection) {
						if teacherDiv.Text() == "" {
							return
						}

						date := h.ToDateWithoutTime(time_)

						var startTime time.Duration
						var endTime time.Duration

						if columnIndex < 5 {
							startTime = p.WeekdaysShift[rowIndex].Start
							endTime = p.WeekdaysShift[rowIndex].End
						} else {
							startTime = p.WeekendsShift[rowIndex].Start
							endTime = p.WeekendsShift[rowIndex].End
						}

						lesson := models.Lesson{
							Id:           primitive.NewObjectID(),
							StudyPlaceId: 0,
							Type:         type_,
							StartDate:    date.Add(startTime),
							EndDate:      date.Add(endTime),
							Subject:      div.Find(".subject").Text(),
							Group:        div.Find(".group").Text(),
							Teacher:      teacherDiv.Text(),
							Room:         div.Find(".place").Text(),
						}

						if models.GetScheduleStateInfoByIndexes(weekIndex, columnIndex, states).State == models.Updated && models.GetScheduleStateInfoByIndexes(weekIndex, columnIndex, p.States).State == models.NotUpdated {
							lessons = append(lessons, &lesson)
						}
					})
				})
			})
			time_ = time_.AddDate(0, 0, -6)
		})
		time_ = time_.AddDate(0, 0, 7)
	})

	p.TempStates = states
	return lessons
}

func (p *KbpParser) GeneralScheduleUpdate(type_ *models.ScheduleTypeInfo) []*models.GeneralLesson {
	var generalLessons []*models.GeneralLesson

	lessons := p.ScheduleUpdate(type_)
	for _, lesson := range lessons {
		if lesson.Type == "ADDED" {
			continue
		}

		weekIndex, _ := lesson.StartDate.ISOWeek()

		generalLesson := models.GeneralLesson{
			Id:           lesson.Id,
			StudyPlaceId: lesson.StudyPlaceId,
			EndTime:      lesson.EndDate.Format("15:04"),
			StartTime:    lesson.StartDate.Format("15:04"),
			Subject:      lesson.Subject,
			Group:        lesson.Group,
			Teacher:      lesson.Teacher,
			Room:         lesson.Room,
			DayIndex:     weekIndex,
			WeekIndex:    lesson.StartDate.Day(),
		}

		generalLessons = append(generalLessons, &generalLesson)
	}

	return generalLessons
}

func (p *KbpParser) ScheduleTypesUpdate() []*models.ScheduleTypeInfo {
	var types []*models.ScheduleTypeInfo

	response, err := http.Get("https://kbp.by/rasp/timetable/view_beta_kbp/?q=")
	if models.BindError(err, 418, h.UNDEFINED).Check() {
		return nil
	}

	if response.StatusCode != 200 {
		models.BindErrorStr("Could not connect to host http://kbp", response.StatusCode, h.UNDEFINED).Check()
		return nil
	}

	document, err := htmlParser.NewDocumentFromReader(response.Body)
	h.CheckError(err, h.WARNING)

	document.Find(".block_back").Find("div").Each(func(ix int, div *htmlParser.Selection) {
		if div.Find("span").Text() == "группа" {
			url, exists := div.Find("a").Attr("href")
			name := div.Find("a").Text()

			if !exists {
				return
			}

			type_ := models.ScheduleTypeInfo{
				ParserAppName: p.GetName(),
				Group:         name,
				Url:           url,
			}

			types = append(types, &type_)
		}
	})
	return types
}

func (p *KbpParser) LoginJournal(user *models.ParseJournalUser) *htmlParser.Document {
	request, _ := http.NewRequest("GET", "https://kbp.by/ej/templates/login_parent.php", nil)
	response, _ := http.DefaultClient.Do(request)
	document, _ := htmlParser.NewDocumentFromReader(response.Body)
	sCode, _ := document.Find("#S_Code").Attr("value")

	cookie := response.Cookies()[0]

	requestString := "action=login_parent&student_name=" + user.Login + "&group_id=" + user.AdditionInfo["groupId"] + "&birth_day=" + user.Password + "&S_Code=" + sCode
	responseBody := bytes.NewBuffer([]byte(requestString))
	request, _ = http.NewRequest("POST", "http://kbp.by/ej/ajax.php", responseBody)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(requestString)))
	request.Header.Add("Host", "kbp.by")
	c := cookie.Name + "=" + cookie.Value
	request.Header.Add("Cookie", c)
	response, _ = http.DefaultClient.Do(request)

	body, _ := ioutil.ReadAll(response.Body)
	if string(body) != "good" {
		return nil
	}

	request, _ = http.NewRequest("GET", "http://kbp.by/ej/templates/parent_journal.php", nil)
	request.Header.Add("Cookie", c)
	request.Header.Add("Host", "kbp.by")
	response, _ = http.DefaultClient.Do(request)

	document, _ = htmlParser.NewDocumentFromReader(response.Body)
	return document
}

func (p *KbpParser) JournalUpdate(user *models.ParseJournalUser) []*models.Mark {
	document := p.LoginJournal(user)
	lessonNames := document.Find("tbody").First().Find(".pupilName").Map(func(i int, selection *htmlParser.Selection) string {
		return strings.TrimSpace(selection.Text())
	})

	marksTable := document.Find("tbody").Last().Find("tr")

	months := map[string]int{
		"январь":   0,
		"февраль":  1,
		"март":     2,
		"апрель":   3,
		"май":      4,
		"июнь":     5,
		"июль":     6,
		"август":   7,
		"сентябрь": 8,
		"октябрь":  9,
		"ноябрь":   10,
		"декабрь":  11,
	}
	var daysAmount [12]int
	marksTable.First().Children().Each(func(i int, selection *htmlParser.Selection) {
		colspan, _ := selection.Attr("colspan")
		days, _ := strconv.Atoi(colspan)

		daysAmount[months[strings.TrimSpace(selection.Text())]] += days
	})

	currentYear := time.Now().Year()
	if time.Now().Month() < 8 {
		currentYear--
	}
	currentMonth := 7
	currentDay := 0

	var dates []time.Time

	marksTable.First().Next().Children().Each(func(i int, selection *htmlParser.Selection) {
		for currentDay >= daysAmount[currentMonth] {
			currentDay = 0
			currentMonth++
			if currentMonth == 12 {
				currentMonth = 0
				currentYear++
			}
		}

		day, _ := strconv.Atoi(selection.Text())
		date := time.Date(currentYear, time.Month(currentMonth+1), day, 0, 0, 0, 0, time.UTC)
		dates = append(dates, date)

		currentDay++
	})

	var marks []*models.Mark

	marksTable.Each(func(rowIndex int, rowSelection *htmlParser.Selection) {
		if rowIndex < 2 {
			return
		}
		rowIndex -= 2

		rowSelection.Children().Each(func(dayIndex int, cellSelection *htmlParser.Selection) {
			if dayIndex >= len(dates) || dates[dayIndex].Before(user.LastParsedDate) {
				return
			}

			cellSelection.Find("span").Each(func(_ int, selection *htmlParser.Selection) {
				mark_ := strings.TrimSpace(selection.Text())
				var lesson models.Lesson
				db.GetLessonByDate(dates[dayIndex], lessonNames[rowIndex], user.AdditionInfo["group"], &lesson)
				if lesson.Id == primitive.NilObjectID {
					return
				}

				mark := models.Mark{
					Id:           primitive.NewObjectID(),
					Mark:         mark_,
					UserId:       user.ID,
					LessonId:     lesson.Id,
					StudyPlaceId: 0,
				}
				marks = append(marks, &mark)
			})
		})
	})

	user.LastParsedDate = dates[len(dates)-1]
	return marks
}

func (p *KbpParser) CommitUpdate() {
	p.States = p.TempStates
	p.TempStates = nil
}

func (p *KbpParser) Init(lesson models.Lesson) {
	var states []*models.ScheduleStateInfo

	date := lesson.StartDate.AddDate(0, 0, -int(lesson.StartDate.Weekday()))
	for len(states) != 14 {
		_, weekIndex := date.ISOWeek()

		state := models.NotUpdated
		if date.Before(lesson.StartDate) {
			state = models.Updated
		}

		states = append(states, &models.ScheduleStateInfo{
			State:     state,
			WeekIndex: weekIndex % 2,
			DayIndex:  int(date.Weekday()),
		})

		date = date.AddDate(0, 0, 1)
	}

	p.States = states

	p.WeekdaysShift = []h.Shift{
		h.BindShift(8, 00, 9, 35),
		h.BindShift(9, 45, 11, 20),
		h.BindShift(11, 50, 13, 25),
		h.BindShift(13, 45, 15, 20),
		h.BindShift(15, 40, 17, 15),
		h.BindShift(17, 25, 19, 0),
		h.BindShift(19, 10, 20, 45),
	}

	p.WeekendsShift = []h.Shift{
		h.BindShift(8, 00, 9, 35),
		h.BindShift(9, 45, 11, 20),
		h.BindShift(11, 30, 13, 5),
		h.BindShift(13, 30, 15, 5),
		h.BindShift(15, 15, 16, 50),
		h.BindShift(17, 0, 18, 35),
		h.BindShift(18, 45, 20, 20),
	}
}
