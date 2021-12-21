package main

import (
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"strings"
)

var KBP = education{
	educationPlaceId:                        0,
	scheduleUpdateCronePattern:              "",
	primaryScheduleUpdateCronePattern:       "",
	scheduleAvailableTypeUpdateCronePattern: "",
	scheduleUpdate:                          UpdateScheduleKbp,
	scheduleStatusUpdate:                    UpdateStateKbp,
	scheduleAvailableTypeUpdate:             UpdateAccessibleTypesKbp,
	availableTypes:                          []string{},
	states:                                  []StateInfo{},
}

func getWeeks(url string) *html.Node {
	resp, err := http.Get("https://kbp.by/rasp/timetable/view_beta_kbp/" + url)
	checkError(err, false)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		checkError(err, false)
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Kbp: Status code %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	checkError(err, false)

	doc, err := html.Parse(strings.NewReader(string(bodyBytes)))
	checkError(err, false)

	return NextSiblings(doc.LastChild.LastChild.FirstChild, 7).FirstChild.NextSibling.LastChild.PrevSibling.FirstChild.NextSibling.LastChild.PrevSibling.FirstChild.NextSibling
}

func UpdateScheduleKbp(url string) []Subject {

	return nil
}

func UpdateStateKbp(url string) []StateInfo {
	weeks := getWeeks(url)

	var states []StateInfo

	weekIndex := 0
	dayIndex := 0

	for week := weeks; week != nil; week = NextSiblings(week, 2) {
		table := week.LastChild.PrevSibling.FirstChild.NextSibling

		statusRow := table.FirstChild.NextSibling.NextSibling
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
	checkError(err, false)

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)

		doc, err := html.Parse(strings.NewReader(bodyString))
		checkError(err, false)

		divs := doc.FirstChild.NextSibling.LastChild.FirstChild.NextSibling.FirstChild.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.FirstChild.NextSibling
		for div := divs; div != nil; div = div.NextSibling.NextSibling {
			for a := div.LastChild.PrevSibling.FirstChild.NextSibling; a != nil; a = a.NextSibling.NextSibling {
				urls = append(urls, a.Attr[0].Val)
			}
		}
	}
	return urls
}
