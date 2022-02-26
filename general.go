package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"
	"time"
)

type StudyPlace struct {
	Id               int32  `bson:"_id"`
	WeeksQuantity    int    `bson:"weeksCount"`
	DaysQuantity     int    `bson:"daysCount"`
	SubjectsQuantity int    `bson:"subjectsCount"`
	Name             string `bson:"name"`
}

func getStudyPlaces(ctx *gin.Context) {
	var res []string

	types, _ := studyPlacesCollection.Find(nil, bson.D{})

	for types.TryNext(nil) {
		res = append(res, "{ \"id\": "+strconv.Itoa(int(types.Current.Lookup("_id").Int32()))+", \"name\": \""+types.Current.Lookup("name").StringValue()+"\"}")
	}

	_, err := fmt.Fprintf(ctx.Writer, "[%s]", strings.Join(res, ", "))
	checkError(err)
}

func getInfo(ctx *gin.Context) {
	var info []gin.H

	for _, studyPlace := range Educations {
		i := gin.H{
			"info":                  studyPlace,
			"isPrimaryCronLaunched": isCronRunning(studyPlace.primaryCron),
			"isGeneralCronLaunched": isCronRunning(studyPlace.generalCron),
		}

		info = append(info, i)
	}

	ctx.JSON(200, gin.H{
		"info":    info,
		"version": 0.1,
		"time":    time.Now(),
	})
}
