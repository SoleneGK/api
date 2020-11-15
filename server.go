package main

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Event struct {
	Id        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Flags     []int     `json:"flags"`
	Data      string    `json:"data"`
}

type EventStore interface {
	GetEventById(id int) Event
}

type Server struct {
	store EventStore
}

const jsonContentType = "application/json"

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/get/"))

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		event := s.store.GetEventById(id)

		if reflect.DeepEqual(event, Event{}) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.Header().Set("content-type", jsonContentType)
			_ = json.NewEncoder(w).Encode(&event)
		}
	}
}
