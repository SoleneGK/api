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
	case http.MethodPost:
		s.registerNewEvent(w, r)
	case http.MethodDelete:
		s.deleteEvent(w)
	default:
		s.rejectRequest(w)
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
 * Status returned:
 *  success: 202 (StatusCreated)
 *  data given can't be registered: 422 (StatusUnprocessableEntity)
 * To be registrable, an event must
 * 	- have a data different from empty string
 *  - have either a timestamp or an author not empty
 */
func (s *Server) registerNewEvent(w http.ResponseWriter, r *http.Request) {
	event := s.getEventFromBodyRequest(r)

	if event.Data == "" || (event.Author == "" && event.Timestamp.IsZero()) {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		s.store.RegisterNewEvent(event)
		w.WriteHeader(http.StatusCreated)
	}

}

func (s *Server) getEventFromBodyRequest(r *http.Request) Event {
	dataSent, _ := ioutil.ReadAll(r.Body)

	event := Event{}
	_ = json.Unmarshal(dataSent, &event)

	return event
}

func (s *Server) isRegistrableEvent(event Event) bool {
	return event.Data != "" && (event.Author != "" || !event.Timestamp.IsZero())
}

func (s *Server) deleteEvent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) rejectRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}
