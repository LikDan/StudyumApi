package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	h "studyum/src/api"
	logApi "studyum/src/api/log"
	"studyum/src/db"
	"studyum/src/models"
	"studyum/src/routes"
	"studyum/src/utils"
	"time"
	//log "github.com/sirupsen/logrus"
	//logApi "studyum/src/api/log"
	//"studyum/src/api/parser"
	//"studyum/src/db"
	//"studyum/src/parser/apps"
	//"studyum/src/routes"
	//"studyum/src/utils"
	//"time"
)

func uptimeHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"message": "hi"})
}

func requestHandler(ctx *gin.Context) {
	response, err := http.Get("https://" + ctx.Query("host"))
	if models.BindError(err, 418, h.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	_, _ = io.Copy(ctx.Writer, response.Body)
}

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	db.Init()

	utils.InitFirebaseApp()
	//parser.InitApps()

	logApi.InitLog()

	r := gin.Default()

	r.HEAD("/api", uptimeHandler)
	r.GET("/request", requestHandler)
	defer logApi.CloseLogFile()

	routes.Bind(r)

	log.Info("Application launched")

	err := r.Run()
	if err != nil {
		log.Fatalf("Error launching server %s", err.Error())
	}
}
