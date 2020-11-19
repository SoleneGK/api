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
}

func main() {
	// Getting env variable from .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	api_port := os.Getenv("API_PORT")

	store = PostGresStore{}

	log.Fatal(http.ListenAndServe(api_port, newServer()))
}
