package kbp

import (
	"bytes"
	htmlParser "github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"strconv"
	"strings"
	"studyum/internal/parser/apps"
	"studyum/internal/parser/dto"
	"studyum/internal/parser/entities"
	"studyum/internal/utils"
	"time"
)

type app struct {
	States     []entities.ScheduleStateInfo
	TempStates []entities.ScheduleStateInfo

	WeekdaysShift []entities.Shift
	WeekendsShift []entities.Shift

	DefaultColor string
	AddedColor   string
	RemovedColor string
}

func NewApp() apps.App {
	states := make([]entities.ScheduleStateInfo, 14)

	for i := 0; i < 7; i++ {
		states[i] = entities.ScheduleStateInfo{
			State:     entities.Updated,
			WeekIndex: 0,
			DayIndex:  i,
		}
	}
	for i := 0; i < 7; i++ {
		states[i] = entities.ScheduleStateInfo{
			State:     entities.NotUpdated,
			WeekIndex: 1,
			DayIndex:  i,
		}
	}

	weekdaysShift := []entities.Shift{
		entities.NewShift(8, 00, 9, 35),
		entities.NewShift(9, 45, 11, 20),
		entities.NewShift(11, 50, 13, 25),
		entities.NewShift(13, 45, 15, 20),
		entities.NewShift(15, 40, 17, 15),
		entities.NewShift(17, 25, 19, 0),
		entities.NewShift(19, 10, 20, 45),
	}

	weekendsShift := []entities.Shift{
		entities.NewShift(8, 00, 9, 35),
		entities.NewShift(9, 45, 11, 20),
		entities.NewShift(11, 30, 13, 5),
		entities.NewShift(13, 30, 15, 5),
		entities.NewShift(15, 15, 16, 50),
		entities.NewShift(17, 0, 18, 35),
		entities.NewShift(18, 45, 20, 20),
	}

	return &app{
		States:        states,
		WeekdaysShift: weekdaysShift,
		WeekendsShift: weekendsShift,

		DefaultColor: "#F1F1F1",
		AddedColor:   "#71AB7F",
		RemovedColor: "#FA6F46",
	}
}

func (a *app) GetName() string              { return "kbp" }
func (a *app) StudyPlaceId() int            { return 0 }
func (a *app) GetUpdateCronPattern() string { return "@every 30m" }

func (a *app) ScheduleUpdate(typeInfo entities.ScheduleTypeInfo) []dto.LessonDTO {
	response, err := http.Get("https://kbp.by/rasp/timetable/view_beta_kbp/" + typeInfo.Url)
	if err != nil {
		return nil
	}

	if response.StatusCode != http.StatusOK {
		return nil
	}

	document, err := htmlParser.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil
	}

	time_ := time.Now().AddDate(0, 0, -int(time.Now().Weekday())).Round(0)

	weeks := document.Find(".find_block").Children().Last().Children()
	if weeks == nil {
		return nil
	}

	var lessons []dto.LessonDTO

	var states []entities.ScheduleStateInfo

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

					stateInfo := entities.ScheduleStateInfo{
						WeekIndex: weekIndex,
						DayIndex:  columnIndex - 1,
					}

					state := strings.TrimSpace(selection.Text())
					if state == "" {
						stateInfo.State = entities.NotUpdated
					} else {
						stateInfo.State = entities.Updated
					}

					states = append(states, stateInfo)
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
					var color string

					if div.HasClass("added") {
						color = a.AddedColor
					} else if div.HasClass("removed") && entities.GetScheduleStateInfoByIndexes(weekIndex, columnIndex, states).State == entities.Updated {
						color = a.RemovedColor
					} else {
						color = a.DefaultColor
					}

					div.Find(".teacher").Each(func(_ int, teacherDiv *htmlParser.Selection) {
						if teacherDiv.Text() == "" {
							return
						}

						var shift entities.Shift

						if columnIndex < 5 {
							shift = a.WeekdaysShift[rowIndex]
						} else {
							shift = a.WeekendsShift[rowIndex]
						}

						shift.Date = utils.ToDateWithoutTime(time_)

						lesson := dto.LessonDTO{
							Shift:        shift,
							PrimaryColor: color,
							Subject:      div.Find(".subject").Text(),
							Group:        div.Find(".group").Text(),
							Teacher:      teacherDiv.Text(),
							Room:         div.Find(".place").Text(),
						}

						if entities.GetScheduleStateInfoByIndexes(weekIndex, columnIndex, states).State == entities.Updated && entities.GetScheduleStateInfoByIndexes(weekIndex, columnIndex, a.States).State == entities.NotUpdated {
							lessons = append(lessons, lesson)
						}
					})
				})
			})
			time_ = time_.AddDate(0, 0, -6)
		})
		time_ = time_.AddDate(0, 0, 7)
	})

	a.TempStates = states
	return lessons
}

func (a *app) GeneralScheduleUpdate(typeInfo entities.ScheduleTypeInfo) []dto.GeneralLessonDTO {
	var generalLessons []dto.GeneralLessonDTO

	lessons := a.ScheduleUpdate(typeInfo)
	for _, lesson := range lessons {
		if lesson.PrimaryColor == a.AddedColor {
			continue
		}

		weekIndex, _ := lesson.Shift.Date.ISOWeek()

		generalLesson := dto.GeneralLessonDTO{
			Shift:     lesson.Shift,
			Subject:   lesson.Subject,
			Group:     lesson.Group,
			Teacher:   lesson.Teacher,
			Room:      lesson.Room,
			WeekIndex: weekIndex,
		}

		generalLessons = append(generalLessons, generalLesson)
	}

	return generalLessons
}

func (a *app) ScheduleTypesUpdate() []entities.ScheduleTypeInfo {
	var types []entities.ScheduleTypeInfo

	response, err := http.Get("https://kbp.by/rasp/timetable/view_beta_kbp/?q=")
	if err != nil {
		return nil
	}

	if response.StatusCode != http.StatusOK {
		return nil
	}

	document, err := htmlParser.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil
	}

	document.Find(".block_back").Find("div").Each(func(ix int, div *htmlParser.Selection) {
		if div.Find("span").Text() == "группа" {
			url, exists := div.Find("a").Attr("href")
			name := div.Find("a").Text()

			if !exists {
				return
			}

			type_ := entities.ScheduleTypeInfo{
				ParserAppName: a.GetName(),
				Group:         name,
				Url:           url,
			}

			types = append(types, type_)
		}
	})
	return types
}

func (a *app) loginJournal(user entities.JournalUser) *htmlParser.Document {
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

	body, _ := io.ReadAll(response.Body)
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

func (a *app) JournalUpdate(user entities.JournalUser) []dto.MarkDTO {
	document := a.loginJournal(user)
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

	var marks []dto.MarkDTO

	marksTable.Each(func(rowIndex int, rowSelection *htmlParser.Selection) {
		if rowIndex < 2 {
			return
		}
		rowIndex -= 2

		rowSelection.Children().Each(func(dayIndex int, cellSelection *htmlParser.Selection) {
			if dayIndex >= len(dates) {
				return
			}

			cellSelection.Find("span").Each(func(_ int, selection *htmlParser.Selection) {
				mark := dto.MarkDTO{
					Mark:       strings.TrimSpace(selection.Text()),
					StudentID:  user.ID,
					LessonDate: dates[dayIndex],
					Subject:    lessonNames[rowIndex],
					Group:      user.AdditionInfo["group"],
				}
				marks = append(marks, mark)
			})
		})
	})

	return marks
}

func (a *app) CommitUpdate() {
	a.States = a.TempStates
	a.TempStates = nil
}
