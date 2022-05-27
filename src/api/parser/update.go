package parser

import (
	"github.com/gin-gonic/gin"
	h "studyum/src/api"
)

func UpdateSchedule(ctx *gin.Context) {
	edu, err := GetEducationViaPasswordRequest(ctx)
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	UpdateDbSchedule(edu)
}
