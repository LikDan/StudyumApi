package parser

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	h "studyium/api"
	"studyium/api/parser/studyPlace"
)

func getEducationViaPasswordRequest(ctx *gin.Context) (*studyPlace.Education, error) {
	password := ctx.Query("password")
	if password == "" {
		return nil, errors.New("provide all params")
	}

	var confirmedEducation *studyPlace.Education

	for _, edu := range Educations {
		if edu.Password == password {
			confirmedEducation = edu
			break
		}
	}

	if confirmedEducation == nil {
		return nil, fmt.Errorf("wrong password")
	}

	return confirmedEducation, nil
}

func StopPrimaryCron(ctx *gin.Context) {
	edu, err := getEducationViaPasswordRequest(ctx)
	if h.CheckError(err) {
		h.ErrorMessage(ctx, err.Error())
		return
	}
	edu.PrimaryCron.Stop()
	edu.LaunchPrimaryCron = false
}

func LaunchPrimaryCron(ctx *gin.Context) {
	edu, err := getEducationViaPasswordRequest(ctx)
	if h.CheckError(err) {
		h.ErrorMessage(ctx, err.Error())
		return
	}
	edu.PrimaryCron.Start()
	edu.LaunchPrimaryCron = true
}
