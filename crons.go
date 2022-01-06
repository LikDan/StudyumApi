package main

import (
	"fmt"
	"net/http"
)

func getEducationViaPasswordRequest(r *http.Request) (*education, error) {
	password, err := getUrlData(r, "password")
	if checkError(err) {
		return nil, err
	}

	var confirmedEducation *education

	for _, edu := range Educations {
		if edu.password == password {
			confirmedEducation = edu
			break
		}
	}

	if confirmedEducation == nil {
		return nil, fmt.Errorf("wrong password")
	}

	return confirmedEducation, nil
}

func stopPrimaryCron(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	edu, err := getEducationViaPasswordRequest(r)
	if checkError(err) {
		_, err := fmt.Fprintln(w, err.Error())
		checkError(err)
		return
	}
	edu.primaryCron.Stop()
	edu.launchPrimaryCron = false
}

func launchPrimaryCron(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	edu, err := getEducationViaPasswordRequest(r)
	if checkError(err) {
		_, err := fmt.Fprintln(w, err.Error())
		checkError(err)
		return
	}
	edu.primaryCron.Start()
	edu.launchPrimaryCron = true
}
