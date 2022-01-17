package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
)

func getEducationViaPasswordRequest(ctx *gin.Context) (*education, error) {
	password := ctx.Query("password")
	if password == "" {
		return nil, errors.New("provide all params")
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

func stopPrimaryCron(ctx *gin.Context) {
	edu, err := getEducationViaPasswordRequest(ctx)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}
	edu.primaryCron.Stop()
	edu.launchPrimaryCron = false
}

func launchPrimaryCron(ctx *gin.Context) {
	edu, err := getEducationViaPasswordRequest(ctx)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}
	edu.primaryCron.Start()
	edu.launchPrimaryCron = true
}
