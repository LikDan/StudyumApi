package main

import (
	"crypto/sha256"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strconv"
)

func getUserFromDb(w http.ResponseWriter, r *http.Request) (bson.M, string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	username, err := getUrlData(r, "username")
	if checkError(err) {
		return nil, buildJSONError("provide all params")
	}
	type_, err := getUrlData(r, "type")
	if checkError(err) {
		return nil, buildJSONError("provide all params")
	}

	if type_ != "password_hash" && type_ != "token" {
		return nil, buildJSONError("provide all params")
	}

	password, err := getUrlData(r, "password")

	if type_ == "password_hash" {
		password = fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	}

	var user bson.M

	userResult := usersCollection.FindOne(nil, bson.M{"username": username, type_: password})
	err = userResult.Decode(&user)
	if checkError(err) {
		return nil, buildJSONError("wrong response")
	}

	if user == nil {
		return nil, buildJSONError("no user")
	}

	return user, ""
}

func createUser(w http.ResponseWriter, r *http.Request) {
	//TODO
}

func editUser(w http.ResponseWriter, r *http.Request) {
	//TODO
}

func getUser(w http.ResponseWriter, r *http.Request) {
	user, stringErr := getUserFromDb(w, r)
	if stringErr != "" {
		_, err := fmt.Fprintln(w, stringErr)
		checkError(err)
		return
	}

	_, err := fmt.Fprintln(w, "{\"username\": \""+user["username"].(string)+
		"\", \"studyPlaceId\": "+strconv.Itoa(int(user["studyPlaceId"].(int32)))+
		", \"type\": \""+user["type"].(string)+
		"\", \"name\": \""+user["name"].(string)+"\"}",
	)
	checkError(err)
}

func getToken(w http.ResponseWriter, r *http.Request) {
	user, stringErr := getUserFromDb(w, r)
	if stringErr != "" {
		_, err := fmt.Fprintln(w, stringErr)
		checkError(err)
		return
	}

	_, err := fmt.Fprintln(w, "{\"token\": \""+user["token"].(string)+"\"}")
	checkError(err)
}

func changeToken(w http.ResponseWriter, r *http.Request) {
	//TODO
}
