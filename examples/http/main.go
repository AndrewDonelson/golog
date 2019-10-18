package main

import (
	"net/http"

	"github.com/AndrewDonelson/golog"
)

type server struct{}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "hello world"}`))
}

func main() {
	log, err := golog.NewLogger(&golog.Options{Module: "dev-http-example"})
	if err != nil {
		panic(err) // Check for error
	}
	log.SetEnvironment(golog.EnvProduction)

	s := &server{}
	http.Handle("/", s)
	log.Print("Listening at localhost:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}
