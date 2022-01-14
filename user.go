package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func getUserFromDbViaCookies(ctx *gin.Context) (bson.M, error) {
	login, err := ctx.Cookie("login")
	if checkError(err) {
		return nil, errors.New("no enough data")
	}

	token, err := ctx.Cookie("password")
	if checkError(err) {
		return nil, errors.New("no enough data")
	}

	var user bson.M

	userResult := usersCollection.FindOne(nil, bson.M{"username": login, "token": token})
	err = userResult.Decode(&user)
	if checkError(err) {
		return nil, errors.New("wrong user or password")
	}

	return user, nil
}

func createUser(ctx *gin.Context) {
	//TODO
}

func editUser(ctx *gin.Context) {
	//TODO
}

func getUserSchedule(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)
	if checkError(err) {
		message(ctx, "error", err.Error(), 418)
	}

	response := gin.H{
		"type":         user["type"].(string),
		"name":         user["name"].(string),
		"studyPlaceId": user["educationPlaceId"].(int32),
	}

	ctx.JSON(200, response)
}

func saveUser(ctx *gin.Context) {
	username := ctx.Query("login")
	password := ctx.Query("password")

	if username == "" || password == "" {
		message(ctx, "error", "provide all params", 418)
		return
	}

	password = fmt.Sprintf("%x", sha256.Sum256([]byte(password)))

	var user bson.M

	userResult := usersCollection.FindOne(nil, bson.M{"username": username, "password_hash": password})
	err := userResult.Decode(&user)

	if checkError(err) {
		message(ctx, "error", "wrong user or password", 418)
		return
	}

	ctx.SetCookie("login", user["username"].(string), 0, "", "", false, false)
	ctx.SetCookie("token", user["token"].(string), 0, "", "", true, false)

	message(ctx, "message", "successful", 200)
}

func deleteUser(ctx *gin.Context) {
	ctx.SetCookie("login", "", -1, "", "", false, false)
	ctx.SetCookie("token", "", -1, "", "", true, false)

	message(ctx, "message", "successful", 200)
}
