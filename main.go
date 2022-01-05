package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"time"
)

func server(w http.ResponseWriter, r *http.Request) {
	log.Println("Trying server access")

	url, err := getUrlData(r, "url")
	checkError(err)

	resp, err := http.Get("http://" + url)
	checkError(err)
	log.Println(resp.StatusCode)
	_, err = fmt.Fprintln(w, resp.StatusCode)
	checkError(err)
}

var generalSubjectsCollection *mongo.Collection
var subjectsCollection *mongo.Collection
var stateCollection *mongo.Collection
var studyPlacesCollection *mongo.Collection

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	log.Println(os.Hostname())

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
