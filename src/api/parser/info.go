package parser

import (
	"github.com/gin-gonic/gin"
	h "studyium/src/api"
	"time"
)

func GetInfo(ctx *gin.Context) {
	var info []gin.H

	for _, studyPlace := range Educations {
		i := gin.H{
			"info":                  studyPlace,
			"isPrimaryCronLaunched": h.IsCronRunning(studyPlace.PrimaryCron),
			"isGeneralCronLaunched": h.IsCronRunning(studyPlace.GeneralCron),
		}

		info = append(info, i)
	}

	ctx.JSON(200, gin.H{
		"info":    info,
		"version": 0.1,
		"time":    time.Now(),
	})
}
