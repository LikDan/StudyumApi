package main

import (
	htmlParser "github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"strings"
)

var KBP = education{
	id:                               0,
	scheduleUpdateCronPattern:        "0 0-59/30 * * * MON-FRI",
	primaryScheduleUpdateCronPattern: "@every 5m",
	primaryCronStartTimePattern:      "0 0 11 * * MON-FRI",
	generalScheduleUpdate:            UpdateGeneralSchedule,
	scheduleUpdate:                   UpdateScheduleKbp,
	scheduleStatesUpdate:             UpdateStateKbp,
	scheduleAvailableTypeUpdate:      UpdateAccessibleTypesKbp,
	availableTypes:                   []string{},
	states:                           []StateInfo{},
	password:                         "kbp-corn-pass",
}

func getWeeks(url string) *html.Node {
	resp, err := http.Get("http://kbp.by/rasp/timetable/view_beta_kbp/" + url)
	checkError(err)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		checkError(err)
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Kbp: Status code %s", resp.Status)
		return nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	checkError(err)

	doc, err := html.Parse(strings.NewReader(string(bodyBytes)))
	checkError(err)

	return NextSiblings(doc.LastChild.LastChild.FirstChild, 7).FirstChild.NextSibling.LastChild.PrevSibling.FirstChild.NextSibling.LastChild.PrevSibling.FirstChild.NextSibling
}

func UpdateScheduleKbp(url string, states []StateInfo) []SubjectFull {
	document, err := htmlParser.NewDocument("http://kbp.by/rasp/timetable/view_beta_kbp/" + url)
	checkError(err)

	weeks := document.Find("tbody")
	if weeks == nil {
		return nil
	}

	var subjects []SubjectFull

	weeks.Each(func(tableIndex int, table *htmlParser.Selection) {
		table.Find("tr").Each(func(rowIndex int, row *htmlParser.Selection) {
			rowIndex -= 2
			row.Find("td").Each(func(columnIndex int, column *htmlParser.Selection) {
				columnIndex -= 1
				column.Find(".pair").Each(func(_ int, div *htmlParser.Selection) {
					var type_ string

					if div.HasClass("added") {
						type_ = "ADDED"
					} else if div.HasClass("removed") && states[tableIndex*6+columnIndex].state == Updated {
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
							weekIndex:        tableIndex,
							type_:            type_,
							educationPlaceId: 0,
						}

						subjects = append(subjects, subject)
					})
				})
			})
		})
	})

	return subjects
}

func UpdateStateKbp(url string) []StateInfo {
	weeks := getWeeks(url)
	if weeks == nil {
		return nil
	}

	var states []StateInfo

	weekIndex := 0
	dayIndex := 0

	for week := weeks; week != nil; week = NextSiblings(week, 2) {
		statusRow := week.LastChild.PrevSibling.FirstChild.NextSibling.FirstChild.NextSibling.NextSibling
		for col := statusRow.FirstChild.NextSibling.NextSibling.NextSibling; col != nil; col = col.NextSibling.NextSibling {
			if col.FirstChild == nil {
				continue
			}

			var state State

			if col.FirstChild.NextSibling == nil && col.FirstChild.Data == "\n\t    \t\t    \t\t\t\t\t" {
				state = NotUpdated
			} else {
				state = Updated
			}

			states = append(states, StateInfo{
				state:            state,
				weekIndex:        weekIndex,
				dayIndex:         dayIndex,
				educationPlaceId: 0,
			})
			dayIndex++
		}
		dayIndex = 0
		weekIndex++
	}

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

func UpdateGeneralSchedule(url string, states []StateInfo) []SubjectFull {
	subjectFull := UpdateScheduleKbp(url, states)

	var subjects []SubjectFull

	for _, subject := range subjectFull {
		if subject.type_ != "ADDED" {
			subjects = append(subjects, subject)
		}
	}

	return subjects
}
