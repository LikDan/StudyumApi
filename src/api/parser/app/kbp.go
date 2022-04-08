package app

import (
	htmlParser "github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	h "studyium/src/api"
	"studyium/src/api/parser/studyPlace"
	"studyium/src/api/schedule"
	"time"
)

var KBP = studyPlace.Education{
	Id:                               0,
	Name:                             "Kbp",
	ScheduleUpdateCronPattern:        "@every 30m",
	PrimaryScheduleUpdateCronPattern: "@every 5m",
	PrimaryCronStartTimePattern:      "0 0 11 * * MON-FRI",
	ScheduleUpdate:                   UpdateScheduleKbp,
	ScheduleStatesUpdate:             UpdateStateKbp,
	ScheduleAvailableTypeUpdate:      UpdateAccessibleTypesKbp,
	AvailableTypes:                   []string{},
	States:                           []schedule.StateInfo{},
	Password:                         "kbp-corn-pass",
}

func UpdateScheduleKbp(url string, states []schedule.StateInfo, oldStates []schedule.StateInfo, isGeneral bool) []schedule.SubjectFull {
	startDurations := []h.Shift{
		h.BindShift(8, 10, 9, 40),
		h.BindShift(9, 50, 11, 20),
		h.BindShift(11, 50, 13, 20),
		h.BindShift(13, 50, 15, 20),
		h.BindShift(15, 40, 17, 10),
		h.BindShift(17, 20, 18, 50),
		h.BindShift(19, 0, 20, 30),
	}

	document, err := htmlParser.NewDocument("http://kbp.by/rasp/timetable/view_beta_kbp/" + url)
	h.CheckError(err, h.WARNING)

	time_ := time.Now().AddDate(0, 0, -int(time.Now().Weekday())).Round(0)

	weeks := document.Find(".find_block").Children().Last().Children()
	if weeks == nil {
		return nil
	}

	var subjects []schedule.SubjectFull

	weeks.Each(func(tableIndex int, table *htmlParser.Selection) {
		var weekIndex int

		if table.Find(".today").First().Text() == "первая неделя" {
			weekIndex = 1
		} else {
			weekIndex = 0
		}

		table.Find("tr").Each(func(rowIndex int, row *htmlParser.Selection) {
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
					} else if div.HasClass("removed") && states[tableIndex*6+columnIndex].State == schedule.Updated {
						type_ = "REMOVED"
					} else {
						type_ = "STAY"
					}

					div.Find(".teacher").Each(func(_ int, teacherDiv *htmlParser.Selection) {
						if teacherDiv.Text() == "" {
							return
						}

						date := h.ToDateWithoutTime(time_)

						subject := schedule.SubjectFull{
							Id:               primitive.NewObjectID(),
							Subject:          div.Find(".subject").Text(),
							Teacher:          teacherDiv.Text(),
							Group:            div.Find(".group").Text(),
							Room:             div.Find(".place").Text(),
							ColumnIndex:      columnIndex,
							RowIndex:         rowIndex,
							WeekIndex:        weekIndex,
							Type_:            type_,
							EducationPlaceId: 0,
							Date:             time_,
							StartTime:        date.Add(startDurations[rowIndex].Start),
							EndTime:          date.Add(startDurations[rowIndex].End),
						}

						if (!isGeneral && states[tableIndex*6+columnIndex].State == schedule.Updated && oldStates[tableIndex*6+columnIndex].State == schedule.NotUpdated) || (isGeneral && type_ != "ADDED") {
							subjects = append(subjects, subject)
						}
					})
				})
			})
			time_ = time_.AddDate(0, 0, -6)
		})
		time_ = time_.AddDate(0, 0, 7)
	})

	return subjects
}

func UpdateStateKbp(url string) []schedule.StateInfo {
	document, err := htmlParser.NewDocument("http://kbp.by/rasp/timetable/view_beta_kbp/" + url)
	h.CheckError(err, h.WARNING)

	var states []schedule.StateInfo

	document.Find(".zamena").Each(func(trIndex int, tr *htmlParser.Selection) {
		tr.Find("th").Each(func(thIndex int, th *htmlParser.Selection) {
			if thIndex == 0 || thIndex > 6 {
				return
			}

			stateInfo := schedule.StateInfo{
				WeekIndex:    trIndex,
				DayIndex:     thIndex - 1,
				StudyPlaceId: 0,
			}

			state := strings.Trim(th.Text(), "\n\t ")
			if state == "" {
				stateInfo.State = schedule.NotUpdated
			} else {
				stateInfo.State = schedule.Updated
			}

			states = append(states, stateInfo)
		})
	})

	return states
}

func UpdateAccessibleTypesKbp() []string {
	var urls []string
	document, err := htmlParser.NewDocument("https://kbp.by/rasp/timetable/view_beta_kbp/?q=")
	h.CheckError(err, h.WARNING)

	document.Find(".block_back").Find("div").Each(func(ix int, div *htmlParser.Selection) {
		if div.Find("span").Text() == "группа" {
			url, exists := div.Find("a").Attr("href")
			if exists {
				urls = append(urls, url)
			}
		}
	})
	return urls
}
