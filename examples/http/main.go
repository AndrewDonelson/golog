package main

import (
	"net/http"

	"github.com/AndrewDonelson/golog"
)

type server struct{}

var log *golog.Logger

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.HandlerLog(w, r)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "hello world"}`))
}

func main() {
	log = golog.NewLogger(&golog.Options{Module: "dev-http-example"})
	log.SetEnvironment(golog.EnvDevelopment)

	s := &server{}
	http.Handle("/", s)
	log.Print("Listening at localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}
