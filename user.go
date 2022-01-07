package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strconv"
)

func createNewUser(username, password string) {
	//TODO
}

func getUserViaToken(w http.ResponseWriter, r *http.Request) {
	username, err := getUrlData(r, "username")
	if checkError(err) {
		return
	}
	token, err := getUrlData(r, "token")
	if checkError(err) {
		return
	}

	var user bson.M

	userResult := usersCollection.FindOne(nil, bson.M{"username": username, "token": token})
	userResult.Decode(&user)

	fmt.Fprintln(w, "{\"username\": \""+user["username"].(string)+
		"\", \"studyPlaceId\": "+strconv.Itoa(int(user["studyPlaceId"].(int32)))+
		", \"type\": \""+user["type"].(string)+
		"\", \"name\": \""+user["name"].(string)+"\"}",
	)
}
