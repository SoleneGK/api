package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	jsonContentType = "application/json"
)

type Event struct {
	Id        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Author    string    `json:"author"`
	Data      string    `json:"data"`
}

type EventStore interface {
	GetEventById(id int) Event
	RegisterNewEvent(event Event)
}

type Server struct {
	store EventStore
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getEvent(w, r)
	case http.MethodPost, http.MethodPatch:
		s.registerNewEvent(w, r)
	}
}

/*
 * getEvent return the event with the id given in parameter as json
 * Status returned:
 * 	success: 200 (StatusOK)
 * 	parameter given is not a number: 422 (StatusUnprocessableEntity)
 *  no event found with given id: 404 (StatusNotFound)
 * In case of success, a json is returned with the data of the event
 */
func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/"))

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		event := s.store.GetEventById(id)

		if reflect.DeepEqual(event, Event{}) {
			w.WriteHeader(http.StatusNotFound)
		}

		w.Header().Set("content-type", jsonContentType)
		_ = json.NewEncoder(w).Encode(&event)
	}
}

/*
 * registerNewEvent create a new event and store it
 * Status returned: 202 (StatusCreated)
 */
func (s *Server) registerNewEvent(w http.ResponseWriter, r *http.Request) {
	dataSent, _ := ioutil.ReadAll(r.Body)

	event := Event{}
	_ = json.Unmarshal(dataSent, &event)

	s.store.RegisterNewEvent(event)
	w.WriteHeader(http.StatusCreated)
}
