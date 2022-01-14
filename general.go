package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"
)

func getStudyPlaces(ctx *gin.Context) {
	var res []string

	types, _ := studyPlacesCollection.Find(nil, bson.D{})

	for types.TryNext(nil) {
		res = append(res, "{ \"id\": "+strconv.Itoa(int(types.Current.Lookup("_id").Int32()))+", \"name\": \""+types.Current.Lookup("name").StringValue()+"\"}")
	}

	_, err := fmt.Fprintf(ctx.Writer, "[%s]", strings.Join(res, ", "))
	checkError(err)
}
