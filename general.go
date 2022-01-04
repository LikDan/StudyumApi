package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strconv"
	"strings"
)

func getStudyPlaces(w http.ResponseWriter, _ *http.Request) {
	var res []string

	types, _ := studyPlacesCollection.Find(nil, bson.D{})

	for types.TryNext(nil) {
		res = append(res, "{ \"id\": "+strconv.Itoa(int(types.Current.Lookup("_id").Int32()))+", \"name\": \""+types.Current.Lookup("name").StringValue()+"\"}")
	}

	_, err := fmt.Fprintf(w, "[%s]", strings.Join(res, ", "))
	checkError(err)
}
