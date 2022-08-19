package apps

import (
	"bytes"
	"context"
	htmlParser "github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"strings"
	"studyum/internal/parser/dto"
	"studyum/internal/parser/entities"
	"studyum/internal/utils"
	"time"
)

type KbpParser struct {
	States     []entities.ScheduleStateInfo
	TempStates []entities.ScheduleStateInfo

	WeekdaysShift []entities.Shift
	WeekendsShift []entities.Shift
}

var KbpApp = KbpParser{}

func (p *KbpParser) GetName() string              { return "kbp" }
func (p *KbpParser) StudyPlaceId() int            { return 0 }
func (p *KbpParser) GetUpdateCronPattern() string { return "@every 30m" }

func (p *KbpParser) ScheduleUpdate(typeInfo entities.ScheduleTypeInfo) []dto.LessonDTO {
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
					var type_ string

					if div.HasClass("added") {
						type_ = "ADDED"
					} else if div.HasClass("removed") && entities.GetScheduleStateInfoByIndexes(weekIndex, columnIndex, states).State == entities.Updated {
						type_ = "REMOVED"
					} else {
						type_ = "STAY"
					}

					div.Find(".teacher").Each(func(_ int, teacherDiv *htmlParser.Selection) {
						if teacherDiv.Text() == "" {
							return
						}

						var shift entities.Shift

						if columnIndex < 5 {
							shift = p.WeekdaysShift[rowIndex]
						} else {
							shift = p.WeekendsShift[rowIndex]
						}

						shift.Date = utils.ToDateWithoutTime(time_)

						lesson := dto.LessonDTO{
							Shift:   shift,
							Type:    type_,
							Subject: div.Find(".subject").Text(),
							Group:   div.Find(".group").Text(),
							Teacher: teacherDiv.Text(),
							Room:    div.Find(".place").Text(),
						}

						if entities.GetScheduleStateInfoByIndexes(weekIndex, columnIndex, states).State == entities.Updated && entities.GetScheduleStateInfoByIndexes(weekIndex, columnIndex, p.States).State == entities.NotUpdated {
							lessons = append(lessons, lesson)
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

func (p *KbpParser) GeneralScheduleUpdate(typeInfo entities.ScheduleTypeInfo) []dto.GeneralLessonDTO {
	var generalLessons []dto.GeneralLessonDTO

	lessons := p.ScheduleUpdate(typeInfo)
	for _, lesson := range lessons {
		if lesson.Type == "ADDED" {
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

func (p *KbpParser) ScheduleTypesUpdate() []entities.ScheduleTypeInfo {
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
				ParserAppName: p.GetName(),
				Group:         name,
				Url:           url,
			}

			types = append(types, type_)
		}
	})
	return types
}

func (p *KbpParser) loginJournal(user entities.JournalUser) *htmlParser.Document {
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

func (p *KbpParser) JournalUpdate(user entities.JournalUser) []dto.MarkDTO {
	document := p.loginJournal(user)
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
			if dayIndex >= len(dates) || dates[dayIndex].Before(user.LastParsedDate) {
				return
			}

			cellSelection.Find("span").Each(func(_ int, selection *htmlParser.Selection) {
				mark := dto.MarkDTO{
					Mark:       strings.TrimSpace(selection.Text()),
					UserId:     user.ID,
					LessonDate: dates[dayIndex],
					Subject:    lessonNames[rowIndex],
					Group:      user.AdditionInfo["group"],
				}
				marks = append(marks, mark)
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

func (p *KbpParser) Init(lesson entities.Lesson) {
	var states []entities.ScheduleStateInfo

	date := lesson.StartDate.AddDate(0, 0, -int(lesson.StartDate.Weekday()))
	for len(states) != 14 {
		_, weekIndex := date.ISOWeek()

		state := entities.NotUpdated
		if date.Before(lesson.StartDate) {
			state = entities.Updated
		}

		states = append(states, entities.ScheduleStateInfo{
			State:     state,
			WeekIndex: weekIndex % 2,
			DayIndex:  int(date.Weekday()),
		})

		date = date.AddDate(0, 0, 1)
	}

	p.States = states

	p.WeekdaysShift = []entities.Shift{
		entities.NewShift(8, 00, 9, 35),
		entities.NewShift(9, 45, 11, 20),
		entities.NewShift(11, 50, 13, 25),
		entities.NewShift(13, 45, 15, 20),
		entities.NewShift(15, 40, 17, 15),
		entities.NewShift(17, 25, 19, 0),
		entities.NewShift(19, 10, 20, 45),
	}

	p.WeekendsShift = []entities.Shift{
		entities.NewShift(8, 00, 9, 35),
		entities.NewShift(9, 45, 11, 20),
		entities.NewShift(11, 30, 13, 5),
		entities.NewShift(13, 30, 15, 5),
		entities.NewShift(15, 15, 16, 50),
		entities.NewShift(17, 0, 18, 35),
		entities.NewShift(18, 45, 20, 20),
	}
}

func (p *KbpParser) OnMarkAdd(_ context.Context, mark entities.Mark, lesson entities.Lesson) map[string]any {
	logrus.Infof("set mark %v, with lesson %v", mark, lesson)
	return nil
}

func (p *KbpParser) OnMarkEdit(_ context.Context, mark entities.Mark, lesson entities.Lesson) map[string]any {
	logrus.Infof("edit mark %v, with lesson %v", mark, lesson)
	return nil
}

func (p *KbpParser) OnMarkDelete(_ context.Context, mark entities.Mark, lesson entities.Lesson) {
	logrus.Infof("delete mark %v, with lesson %v", mark, lesson)
}

func (p *KbpParser) OnLessonAdd(_ context.Context, lesson entities.Lesson) map[string]any {
	logrus.Infof("add lesson %v", lesson)
	return nil
}

func (p *KbpParser) OnLessonEdit(_ context.Context, lesson entities.Lesson) map[string]any {
	logrus.Infof("add edit %v", lesson)
	return nil
}

func (p *KbpParser) OnLessonDelete(_ context.Context, lesson entities.Lesson) {
	logrus.Infof("add delete %v", lesson)
}
