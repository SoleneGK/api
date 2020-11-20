package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var store interface {
	GetEventById(id int) Event
	GetAllEvents() []Event
	GetEventsByFlag(flag int) []Event
	RegisterNewEvents(eventList []Event) int
}

func main() {
	// Getting env variable from .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	api_port := os.Getenv("API_PORT")

	store = PostGreStore{}

	log.Fatal(http.ListenAndServe(api_port, newServer()))
}
