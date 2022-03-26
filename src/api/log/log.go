package log

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
	h "studyium/src/api"
	"time"
)

type Log struct {
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
	Time  time.Time `json:"time" time_format:"2006-01-02"`
}

var logFile *os.File
var ginWriter *GinWriter

func getLog(ctx *gin.Context) {
	startTime, _ := time.Parse("2006-01-02T15:04:05", ctx.Query("startTime"))
	endTime, _ := time.Parse("2006-01-02T15:04:05", ctx.Query("endTime"))

	if !startTime.IsZero() {
		startTime = startTime.Add(time.Hour * -3)
	}
	if !endTime.IsZero() {
		endTime = endTime.Add(time.Hour * -3)
	}

	r, err := os.Open(logFile.Name())
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

func deleteLog(*gin.Context) {
	CloseLogFile()

	file, err := os.OpenFile("logs.jsonl", os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
	if !h.CheckError(err, h.WARNING) {
		logFile = file
		log.SetOutput(file)
	}

	file, err = os.OpenFile("requests.jsonl", os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
	if !h.CheckError(err, h.WARNING) {
		ginWriter.File = file
	}

	log.Warning("Logs has been truncated")
}

func CloseLogFile() {
	log.Warning("Application closed")
	_ = logFile.Close()
	_ = ginWriter.File.Close()
}

func InitLog() {
	log.SetFormatter(&log.JSONFormatter{})
	file, err := os.OpenFile("logs.jsonl", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if !h.CheckError(err, h.WARNING) {
		logFile = file
		log.SetOutput(file)
	}

	file, err = os.OpenFile("requests.jsonl", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if !h.CheckError(err, h.WARNING) {
		ginWriter = &GinWriter{
			File: file,
		}

		gin.DefaultWriter = ginWriter
	}
}

func BuildRequests(api *gin.RouterGroup) {
	api.GET("", getLog)
	api.DELETE("", deleteLog)

	api.GET("/requests", getRequests)
}
