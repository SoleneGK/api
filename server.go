package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	jsonContentType       = "application/json"
	api_url               = "/api/game-event/"
	lineNumberResponseKey = "affectedlines"
)

type Event struct {
	Id        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Flags     []int     `json:"flags"`
	Data      string    `json:"data"`
}

func newServer() http.Handler {
	router := mux.NewRouter()

	// GET requests
	router.HandleFunc(api_url+"{id}", getByIdHandler).Methods(http.MethodGet)
	router.HandleFunc(api_url, getAllHandler).Methods(http.MethodGet)
	router.HandleFunc(api_url+"getFlag/{id}", getByFlagHandler).Methods(http.MethodGet)

	// POST request
	router.HandleFunc(api_url, postHandler).Methods(http.MethodPost)

	return router
}

func getByIdHandler(w http.ResponseWriter, r *http.Request) {
	id, err := extractIntFromURL(r, api_url)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		event := store.GetEventById(id)
		sendEvent(event, w)
	}
}

func getAllHandler(w http.ResponseWriter, r *http.Request) {
	listEvent := store.GetAllEvents()
	writeResponseBody(w, &listEvent)
}

func getByFlagHandler(w http.ResponseWriter, r *http.Request) {
	flag, err := extractIntFromURL(r, api_url+"getFlag/")

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		eventList := store.GetEventsByFlag(flag)
		sentEventList(eventList, w)
	}
}

func extractIntFromURL(r *http.Request, prefix string) (int, error) {
	return strconv.Atoi(strings.TrimPrefix(r.URL.Path, prefix))
}

func sendEvent(event Event, w http.ResponseWriter) {
	if isEmptyEvent(event) {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Header().Set("content-type", jsonContentType)
		writeResponseBody(w, &event)
	}
}

func isEmptyEvent(event Event) bool {
	return reflect.DeepEqual(event, Event{})
}

func sentEventList(eventList []Event, w http.ResponseWriter) {
	if isEmptyEventList(eventList) {
		w.WriteHeader(http.StatusNotFound)
	} else {
		writeResponseBody(w, &eventList)
	}
}

func isEmptyEventList(eventList []Event) bool {
	return len(eventList) == 0
}

func writeResponseBody(w http.ResponseWriter, content interface{}) {
	_ = json.NewEncoder(w).Encode(content)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	eventList := getEventListFromRequest(r)

	linesAffected := store.RegisterNewEvents(eventList)

	_, _ = w.Write(formatLineNumberResponse(linesAffected))
}

func getEventListFromRequest(r *http.Request) []Event {
	dataSent, _ := ioutil.ReadAll(r.Body)

	eventList := []Event{}
	_ = json.Unmarshal(dataSent, &eventList)

	return eventList
}

func formatLineNumberResponse(lineNumber int) []byte {
	responseAsString := fmt.Sprintf("{\"%s\":%d}", lineNumberResponseKey, 2)
	responseAsBytes := []byte(responseAsString)
	return responseAsBytes
}
