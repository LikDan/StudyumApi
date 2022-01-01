package main

import (
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
	scheduleUpdate:                   UpdateScheduleKbp,
	scheduleStatesUpdate:             UpdateStateKbp,
	scheduleAvailableTypeUpdate:      UpdateAccessibleTypesKbp,
	availableTypes:                   []string{},
	states:                           []StateInfo{},
}

func getWeeks(url string) *html.Node {
	resp, err := http.Get("https://kbp.by/rasp/timetable/view_beta_kbp/" + url)
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
	weeks := getWeeks(url)
	if weeks == nil {
		return nil
	}

	var subjects []SubjectFull

	weekIndex := 0
	rowIndex := 0
	columnIndex := 0
	for week := weeks; week != nil; week = NextSiblings(week, 2) {
		for c := week.LastChild.PrevSibling.FirstChild.NextSibling.FirstChild.NextSibling.NextSibling.NextSibling.NextSibling; c != nil; c = c.NextSibling.NextSibling {
			for i := c.FirstChild.NextSibling.NextSibling.NextSibling; i != nil; i = i.NextSibling.NextSibling {
				addSubject := func(subjectName, teacher, room, group, type_ string) {
					subject := SubjectFull{
						subject:          normalizeStr(subjectName),
						teacher:          normalizeStr(teacher),
						group:            normalizeStr(group),
						room:             normalizeStr(room),
						columnIndex:      columnIndex,
						rowIndex:         rowIndex,
						weekIndex:        weekIndex,
						type_:            type_,
						educationPlaceId: 0,
					}
					for _, s := range subjects {
						if s == subject {
							return
						}
					}

					subjects = append(subjects, subject)
				}

				for div := i.FirstChild; div != nil; div = div.NextSibling {
					if div.Data == "div" {
						if strings.Contains(div.Attr[0].Val, "empty-pair") {
							continue
						}
						subject := div.FirstChild.NextSibling.FirstChild.NextSibling.FirstChild.FirstChild.Data
						teacher := ""
						teacherDiv := div.FirstChild.NextSibling.FirstChild.NextSibling.NextSibling.NextSibling.FirstChild.FirstChild
						if teacherDiv != nil {
							teacher = teacherDiv.Data
						}
						room := div.FirstChild.NextSibling.NextSibling.NextSibling.LastChild.PrevSibling.FirstChild.FirstChild.Data
						group := div.FirstChild.NextSibling.NextSibling.NextSibling.FirstChild.NextSibling.FirstChild.FirstChild.FirstChild.Data
						teacher2Node := div.FirstChild.NextSibling.LastChild.PrevSibling.FirstChild.FirstChild

						if strings.Contains(div.Attr[0].Val, "added") {
							addSubject(subject, teacher, room, group, "ADDED")
							if teacher2Node != nil {
								addSubject(subject, teacher2Node.Data, room, group, "ADDED")
							}
						} else if strings.Contains(div.Attr[0].Val, "removed") &&
							states[weekIndex*6+columnIndex].state != NotUpdated {
							addSubject(subject, teacher, room, group, "REMOVED")
							if teacher2Node != nil {
								addSubject(subject, teacher2Node.Data, room, group, "REMOVED")
							}
						} else {
							addSubject(subject, teacher, room, group, "STAY")
							if teacher2Node != nil {
								addSubject(subject, teacher2Node.Data, room, group, "STAY")
							}
						}
					}
				}
				columnIndex++
			}
			rowIndex++
			columnIndex = 0
		}
		rowIndex = 0
		columnIndex = 0
		weekIndex++
	}

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
	resp, err := http.Get("https://kbp.by/rasp/timetable/view_beta_kbp/")
	checkError(err)

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)

		doc, err := html.Parse(strings.NewReader(bodyString))
		checkError(err)

		divs := doc.FirstChild.NextSibling.LastChild.FirstChild.NextSibling.FirstChild.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.FirstChild.NextSibling
		for div := divs; div != nil; div = div.NextSibling.NextSibling {
			for a := div.LastChild.PrevSibling.FirstChild.NextSibling; a != nil; a = a.NextSibling.NextSibling {
				urls = append(urls, a.Attr[0].Val)
			}
		}
	}
	return urls
}
