package main

import (
	"github.com/robfig/cron"
	"net/http"
	"os"
)

func server(http.ResponseWriter, *http.Request) {

}

func main() {
	Launch()

	serverCrone := cron.New()
	err := serverCrone.AddFunc("@every 15m", func() {
		_, err := http.Get("http://kbp-server.herokuapp.com/")
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
