package schedule

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"
	h "studyium/api"
	"studyium/db"
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

	types, _ := db.StudyPlacesCollection.Find(nil, bson.D{})

	for types.TryNext(nil) {
		res = append(res, "{ \"id\": "+strconv.Itoa(int(types.Current.Lookup("_id").Int32()))+", \"name\": \""+types.Current.Lookup("name").StringValue()+"\"}")
	}

	_, err := fmt.Fprintf(ctx.Writer, "[%s]", strings.Join(res, ", "))
	h.CheckError(err)
}
