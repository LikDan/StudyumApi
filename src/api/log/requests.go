package log

import (
	"github.com/gin-gonic/gin"
	"os"
	h "studyum/src/api"
)

func getRequests(ctx *gin.Context) {
	r, err := os.Open(ginWriter.File.Name())
	if err != nil {
		print(err.Error())
	}

	var logs []RequestInfo
	h.DecodeJsonLines(r, &logs)

	ctx.JSON(200, logs)
}
