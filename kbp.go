package main

import (
	htmlParser "github.com/PuerkitoBio/goquery"
	"strings"
	"time"
)

var KBP = education{
	id:                               0,
	scheduleUpdateCronPattern:        "@every 30m",
	primaryScheduleUpdateCronPattern: "@every 5m",
	primaryCronStartTimePattern:      "0 0 11 * * MON-FRI",
	scheduleUpdate:                   UpdateScheduleKbp,
	scheduleStatesUpdate:             UpdateStateKbp,
	scheduleAvailableTypeUpdate:      UpdateAccessibleTypesKbp,
	availableTypes:                   []string{},
	states:                           []StateInfo{},
	password:                         "kbp-corn-pass",
}

func UpdateScheduleKbp(url string, states []StateInfo, isGeneral bool) []SubjectFull {
	document, err := htmlParser.NewDocument("http://kbp.by/rasp/timetable/view_beta_kbp/" + url)
	checkError(err)

	time_ := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))

	weeks := document.Find(".find_block").Children().Last().Children()
	if weeks == nil {
		return nil
	}

	var subjects []SubjectFull

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
					} else if div.HasClass("removed") && states[tableIndex*6+columnIndex].State == Updated {
						type_ = "REMOVED"
					} else {
						type_ = "STAY"
					}

					div.Find(".teacher").Each(func(_ int, teacherDiv *htmlParser.Selection) {
						if teacherDiv.Text() == "" {
							return
						}

						subject := SubjectFull{
							subject:          div.Find(".subject").Text(),
							teacher:          teacherDiv.Text(),
							group:            div.Find(".group").Text(),
							room:             div.Find(".place").Text(),
							columnIndex:      columnIndex,
							rowIndex:         rowIndex,
							weekIndex:        weekIndex,
							type_:            type_,
							educationPlaceId: 0,
							date:             time_,
						}

						if (!isGeneral && states[tableIndex*6+columnIndex].State == Updated) || (isGeneral && type_ != "ADDED") {
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

func UpdateStateKbp(url string) []StateInfo {
	document, err := htmlParser.NewDocument("http://kbp.by/rasp/timetable/view_beta_kbp/" + url)
	checkError(err)

	var states []StateInfo

	document.Find(".zamena").Each(func(trIndex int, tr *htmlParser.Selection) {
		tr.Find("th").Each(func(thIndex int, th *htmlParser.Selection) {
			if thIndex == 0 || thIndex > 6 {
				return
			}

			stateInfo := StateInfo{
				WeekIndex:    trIndex,
				DayIndex:     thIndex - 1,
				StudyPlaceId: 0,
			}

			state := strings.Trim(th.Text(), "\n\t ")
			if state == "" {
				stateInfo.State = NotUpdated
			} else {
				stateInfo.State = Updated
			}

			states = append(states, stateInfo)
		})
	})

	return states
}

func UpdateAccessibleTypesKbp() []string {
	var urls []string
	document, err := htmlParser.NewDocument("https://kbp.by/rasp/timetable/view_beta_kbp/?q=")
	checkError(err)

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
