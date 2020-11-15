package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Getting env variable from .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	server := &Server{}
	api_port := os.Getenv("API_PORT")

	if err := http.ListenAndServe(api_port, server); err != nil {
		log.Fatalf("could not listen on port %s %v", api_port, err)
	}
}
