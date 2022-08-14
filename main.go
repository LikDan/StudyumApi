package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"studyum/src/db"
	"studyum/src/models"
	"studyum/src/parser"
	"studyum/src/routes"
	"studyum/src/utils"
	"time"
)

func uptimeHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"message": "hi"})
}

func requestHandler(ctx *gin.Context) {
	response, err := http.Get("https://" + ctx.Query("host"))
	if models.BindError(err, 418, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	_, _ = io.Copy(ctx.Writer, response.Body)
}

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	db.Init()

	utils.InitFirebaseApp()
	parser.InitApps()

	r := gin.Default()

	r.HEAD("/api", uptimeHandler)
	r.GET("/request", requestHandler)

	routes.Bind(r)

	log.Info("Application launched")

	err := r.Run()
	if err != nil {
		log.Fatalf("Error launching server %s", err.Error())
	}
}
