package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
)

func getUserFromDbViaCookies(ctx *gin.Context) (*User, error) {
	login, loginErr := ctx.Cookie("login")
	token, tokenErr := ctx.Cookie("token")

	if checkError(loginErr) || checkError(tokenErr) {
		return nil, errors.New("not authorized")
	}

	var user User

	userResult := usersCollection.FindOne(nil, bson.M{"login": login, "token": token})
	err := userResult.Decode(&user)
	if checkError(err) {
		return nil, errors.New("not authorized")
	}

	return &user, nil
}

func getLogin(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	message(ctx, "login", user.Login, 200)
}

func createUser(ctx *gin.Context) {
	login := ctx.Query("login")
	password := ctx.Query("password")

	type_ := ctx.Query("type")
	name := ctx.Query("name")
	studyPlaceId := ctx.Query("studyPlaceId")

	stay, err := strconv.ParseBool(ctx.DefaultQuery("stay", "false"))

	if login == "" || password == "" || type_ == "" || name == "" || studyPlaceId == "" || len(password) < 8 {
		errorMessage(ctx, "provide all params")
		return
	}

	if err != nil {
		errorMessage(ctx, "not valid params")
		return
	}

	password = fmt.Sprintf("%x", sha256.Sum256([]byte(password)))

	_, err = usersCollection.InsertOne(nil, bson.D{{"login", login}, {"password_hash", password}, {"type", type_}, {"name", name}, {"studyPlaceId", studyPlaceId}})
	if err != nil {
		errorMessage(ctx, err.Error())
		return
	}

	if stay {
		loginUser(ctx)
	}

	message(ctx, "message", "successful", 200)
}

func editUser(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	type_ := ctx.DefaultQuery("type", user.Type)
	name := ctx.DefaultQuery("name", user.Name)
	studyPlaceId, err := strconv.Atoi(ctx.DefaultQuery("studyPlaceId", strconv.Itoa(user.StudyPlaceId)))

	if err != nil {
		errorMessage(ctx, "not valid params")
		return
	}

	_, err = usersCollection.UpdateByID(nil, user.Id, bson.D{{"$set", bson.D{{"type", type_}, {"name", name}, {"studyPlaceId", studyPlaceId}}}})
	if err != nil {
		errorMessage(ctx, err.Error())
		return
	}

	message(ctx, "message", "successful", 200)
}

func deleteUser(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)
	if err != nil {
		errorMessage(ctx, err.Error())
		return
	}

	_, err = usersCollection.DeleteOne(nil, user)
	if err != nil {
		errorMessage(ctx, err.Error())
		return
	}

	logoutUser(ctx)
}

func loginUser(ctx *gin.Context) {
	login := ctx.Query("login")
	password := ctx.Query("password")

	if login == "" || password == "" {
		errorMessage(ctx, "provide all params")
		return
	}

	password = fmt.Sprintf("%x", sha256.Sum256([]byte(password)))

	var user bson.M

	userResult := usersCollection.FindOne(nil, bson.M{"login": login, "password_hash": password})
	err := userResult.Decode(&user)

	if checkError(err) {
		errorMessage(ctx, "wrong user or password")
		return
	}

	ctx.SetCookie("login", user["login"].(string), 0, "", "", false, false)
	ctx.SetCookie("token", user["token"].(string), 0, "", "", true, false)

	message(ctx, "message", "successful", 200)
}

func logoutUser(ctx *gin.Context) {
	ctx.SetCookie("login", "", -1, "", "", false, false)
	ctx.SetCookie("token", "", -1, "", "", true, false)

	message(ctx, "message", "successful", 200)
}

func getUserInfo(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)
	if checkError(err) {
		return
	}

	ctx.JSON(200, user)
}

type User struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	Login        string             `json:"login"`
	Type         string             `json:"type" bson:"type"`
	Name         string             `json:"name"`
	FullName     string             `json:"fullName"`
	Permissions  []string           `json:"permissions"`
	StudyPlaceId int                `json:"studyPlaceId"`
}
