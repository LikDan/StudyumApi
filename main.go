package main

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var generalSubjectsCollection *mongo.Collection
var subjectsCollection *mongo.Collection
var stateCollection *mongo.Collection
var studyPlacesCollection *mongo.Collection
var usersCollection *mongo.Collection

func indexHandler(ctx *gin.Context) {
	message(ctx, "message", "hi", 200)
}

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	client, err := mongo.NewClient(options.Client().ApplyURI(getDbUrl()))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(nil)
	checkError(err)

	studyPlacesCollection = client.Database("General").Collection("StudyPlaces")
	usersCollection = client.Database("General").Collection("Users")

	subjectsCollection = client.Database("Schedule").Collection("Subjects")
	generalSubjectsCollection = client.Database("Schedule").Collection("General")
	stateCollection = client.Database("Schedule").Collection("States")

	initFirebaseApp()
	Launch()

	r := gin.Default()

	r.GET("/api", indexHandler)

	api := r.Group("/api")

	userGroup := api.Group("/user")

	api.GET("/schedule", getSchedule)
	api.GET("/schedule/types", getScheduleTypes)
	api.GET("/schedule/update", updateSchedule)

	userGroup.GET("/login", loginUser)
	userGroup.GET("/logout", logoutUser)
	userGroup.GET("/edit", editUser)
	userGroup.GET("/create", createUser)
	userGroup.GET("/delete", deleteUser)
	userGroup.GET("/getLogin", getLogin)

	userGroup.GET("/schedule", getUserSchedule)

	api.GET("/studyPlaces", getStudyPlaces)
	api.GET("/info", getInfo)

	api.GET("/stopPrimaryUpdates", stopPrimaryCron)
	api.GET("/launchPrimaryUpdates", launchPrimaryCron)

	err = r.Run()
	checkError(err)
}
