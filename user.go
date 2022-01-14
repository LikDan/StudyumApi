package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strconv"
	"strings"
)

func getUserFromDb(ctx *gin.Context) (bson.M, error) {
	username := ctx.Query("username")
	type_ := ctx.Query("type")
	password := ctx.Query("password")

	if username == "" || type_ == "" || password == "" {
		return nil, errors.New("provide all params")
	}

	if type_ != "password_hash" && type_ != "token" {
		return nil, errors.New("wrong type")
	}

	if type_ == "password_hash" {
		password = fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	}

	var user bson.M

	userResult := usersCollection.FindOne(nil, bson.M{"username": username, type_: password})
	err := userResult.Decode(&user)
	if checkError(err) {
		return nil, errors.New("wrong user or password")
	}

	return user, nil
}

func createUser(w http.ResponseWriter, r *http.Request) {
	//TODO
}

func editUser(w http.ResponseWriter, r *http.Request) {
	//TODO
}

func getUser(ctx *gin.Context) {
	user, err := getUserFromDb(ctx)
	if err != nil {
		message(ctx, "error", err.Error(), 418)
		return
	}

	var rights []string

	for _, right := range user["rights"].(primitive.A) {
		rights = append(rights, right.(string))
	}

	_, err = fmt.Fprintln(ctx.Writer, "{\"username\": \""+user["username"].(string)+
		"\", \"studyPlaceId\": "+strconv.Itoa(int(user["studyPlaceId"].(int32)))+
		", \"type\": \""+user["type"].(string)+
		"\", \"name\": \""+user["name"].(string)+
		"\", \"rights\": [\""+strings.Join(rights, "\", \"")+"\"]}",
	)
	checkError(err)
}

func getToken(ctx *gin.Context) {
	user, err := getUserFromDb(ctx)
	if err != nil {
		message(ctx, "error", err.Error(), 418)
		return
	}

	_, err = fmt.Fprintln(ctx.Writer, "{\"token\": \""+user["token"].(string)+"\"}")
	checkError(err)
}

func changeToken(w http.ResponseWriter, r *http.Request) {
	//TODO
}
