package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"time"
)

func server(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		return
	}
	url, err := getUrlData(r, "url")
	checkError(err)

	resp, err := http.Get("https://" + url)
	if err != nil {
		_, err = fmt.Fprintln(w, "No such host")
		checkError(err)
		return
	}

	_, err = fmt.Fprintln(w, resp.StatusCode)
	checkError(err)
}

var generalSubjectsCollection *mongo.Collection
var subjectsCollection *mongo.Collection
var stateCollection *mongo.Collection
var studyPlacesCollection *mongo.Collection
var usersCollection *mongo.Collection

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	serverCron := cron.New()
	err := serverCron.AddFunc("@every 10m", func() {
		_, err := http.Get("https://collegehelper-3f572.appspot.com/")
		checkError(err)
	})
	checkError(err)
	serverCron.Start()

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

	router := mux.NewRouter()

	router.HandleFunc("/schedule", getSchedule)
	router.HandleFunc("/schedule/types", getScheduleTypes)
	router.HandleFunc("/schedule/update", updateSchedule)

	router.HandleFunc("/studyPlaces", getStudyPlaces)
	router.HandleFunc("/getUser", getUser)
	router.HandleFunc("/getToken", getToken)

	router.HandleFunc("/stopPrimaryUpdates", stopPrimaryCron)
	router.HandleFunc("/launchPrimaryUpdates", launchPrimaryCron)

	router.HandleFunc("/", server)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		println("Port set to 8080")
	}
	err = http.ListenAndServe(":"+port, router)
	checkError(err)
}
