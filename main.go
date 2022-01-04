package main

import (
	"fmt"
	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
)

func server(w http.ResponseWriter, _ *http.Request) {
	log.Println("Trying server access")
	_, err := fmt.Fprintln(w, "College Helper")
	checkError(err)
}

var generalSubjectsCollection *mongo.Collection
var subjectsCollection *mongo.Collection
var stateCollection *mongo.Collection
var studyPlacesCollection *mongo.Collection

func main() {
	_, err := http.Get("http://kbp.by/rasp/timetable/view_beta_kbp/")
	log.Println(err)

	client, err := mongo.NewClient(options.Client().ApplyURI(getDbUrl()))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(nil)
	checkError(err)

	studyPlacesCollection = client.Database("General").Collection("StudyPlaces")
	subjectsCollection = client.Database("Schedule").Collection("Subjects")
	generalSubjectsCollection = client.Database("Schedule").Collection("General")
	stateCollection = client.Database("Schedule").Collection("States")

	initFirebaseApp()
	Launch()

	serverCrone := cron.New()
	err = serverCrone.AddFunc("@every 10m", func() {
		_, err := http.Get("https://college-helper.herokuapp.com/")
		checkError(err)
	})
	checkError(err)
	serverCrone.Start()

	http.HandleFunc("/schedule", getSchedule)
	http.HandleFunc("/schedule/types", getScheduleTypes)
	http.HandleFunc("/studyPlaces", getStudyPlaces)
	http.HandleFunc("/", server)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		println("Port set to 8080")
	}
	err = http.ListenAndServe(":"+port, nil)
	checkError(err)
}
