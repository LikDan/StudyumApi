package schedule

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	h "studyium/api"
	"studyium/db"
)

type StudyPlace struct {
	Id               int32  `json:"id" bson:"_id"`
	WeeksQuantity    int    `json:"weeksQuantity" bson:"weeksCount"`
	DaysQuantity     int    `json:"daysQuantity" bson:"daysCount"`
	SubjectsQuantity int    `json:"subjectsQuantity" bson:"subjectsCount"`
	Name             string `json:"name" bson:"name"`
}

func getStudyPlaces(ctx *gin.Context) {
	var places []StudyPlace

	types, _ := db.StudyPlacesCollection.Find(nil, bson.D{})
	err := types.All(nil, &places)
	if h.CheckError(err) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	ctx.JSON(200, places)
}
