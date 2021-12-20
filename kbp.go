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
	scheduleAvailableTypeUpdate:             UpdateAccessibleTypesKbp,
	availableTypes:                          []string{},
}

func UpdateScheduleKbp() {

}

func UpdateStateKbp() {

}

func UpdateAccessibleTypesKbp() []string {
	var scheduleUrls []string
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
				scheduleUrls = append(scheduleUrls, a.Attr[0].Val)
			}
		}
	}
	return scheduleUrls
}
