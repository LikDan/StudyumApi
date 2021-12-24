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
	checkError(err, false)
}

var subjectsCollection *mongo.Collection
var stateCollection *mongo.Collection

func main() {
	dbUrl := "mongodb+srv://likdan:Byd7FhSBtNfdaJ7w@maincluster.nafh0.mongodb.net/main?retryWrites=true&w=majority"

	client, err := mongo.NewClient(options.Client().ApplyURI(dbUrl))
	if err != nil {
		log.Fatal(err)
	}

	subjectsCollection = client.Database("main").Collection("Schedule")
	stateCollection = client.Database("main").Collection("Replacement")

	initFirebaseApp()
	Launch()

	serverCrone := cron.New()
	err = serverCrone.AddFunc("@every 10m", func() {
		_, err := http.Get("https://college-helper.herokuapp.com/")
		checkError(err, true)
	})
	checkError(err, true)
	serverCrone.Start()

	http.HandleFunc("/", server)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		println("Port set to 8080")
	}
	err = http.ListenAndServe(":"+port, nil)
	checkError(err, true)
}
