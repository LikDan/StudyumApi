package log

import (
	"github.com/gin-gonic/gin"
	"os"
	h "studyium/api"
	"time"
)

type Log struct {
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
	Time  time.Time `json:"time" time_format:"2006-01-02"`
}

func getLog(ctx *gin.Context) {
	startTime, sErr := time.Parse("2006-01-02T15:04:05", ctx.Query("startTime"))
	endTime, eErr := time.Parse("2006-01-02T15:04:05", ctx.Query("endTime"))

	if sErr != nil || eErr != nil {
		print("error")
	}

	if !startTime.IsZero() {
		startTime = startTime.Add(time.Hour * -3)
	}
	if !endTime.IsZero() {
		endTime = endTime.Add(time.Hour * -3)
	}

	r, err := os.Open("logs.jsonl")
	if err != nil {
		print(err.Error())
	}

	var allLogs []Log
	h.DecodeJsonLines(r, &allLogs)

	var logs []Log

	for _, l := range allLogs {
		if (startTime.IsZero() || l.Time.After(startTime)) && (endTime.IsZero() || l.Time.Before(endTime)) {
			logs = append(logs, l)
		}
	}
	ctx.JSON(200, logs)
}

func deleteLog(ctx *gin.Context) {
	err := os.Remove("log.jsonl")
	if err != nil {
		return
	}
}

func BuildRequests(api *gin.RouterGroup) {
	api.GET("", getLog)

	api.DELETE("", deleteLog)
}
