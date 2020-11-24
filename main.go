package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var store interface {
	GetEventById(id int) Event
	GetAllEvents() []Event
	GetEventsByFlag(flag int) []Event
	RegisterNewEvents(eventList []Event) int
}

var clock interface {
	Now() time.Time
}

type RealClock struct{}

func (c RealClock) Now() time.Time {
	return time.Now()
}

func main() {
	// Getting env variable from .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	api_port := os.Getenv("API_PORT")

	store = PostGreStore{}
	clock = RealClock{}

	log.Fatal(http.ListenAndServe(api_port, newServer()))
}
