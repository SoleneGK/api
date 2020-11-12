package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestReadRequest(t *testing.T) {
	events := []Event{
		Event{
			Id:        1,
			Timestamp: time.Unix(1605107095, 0),
			Author:    "Solène",
		},
		Event{
			Id:        2,
			Timestamp: time.Unix(1605107099, 0),
			Author:    "Camille",
		},
	}

	store := StubEventStore{events: events}
	server := &Server{&store}

	t.Run("return event with id 1", func(t *testing.T) {
		request := newGetIdRequest(1)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := getEventFromResponse(t, response.Body)

		assertStatus(t, response.Code, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertEvent(t, got, events[0])
	})

	t.Run("return event with id 2", func(t *testing.T) {
		request := newGetIdRequest(2)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := getEventFromResponse(t, response.Body)
		assertStatus(t, response.Code, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertEvent(t, got, events[1])
	})

	t.Run("return a 404 when no event with id exists", func(t *testing.T) {
		request := newGetIdRequest(5)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("return an error when id given is not a number", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/aaa", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnprocessableEntity)
	})
}

func TestRegisterRequest(t *testing.T) {
	event := Event{
		Timestamp: time.Unix(1605107095, 0),
		Author:    "Solène",
	}

	store := StubEventStore{
		events:        []Event{event},
		registerCalls: nil,
	}
	server := &Server{&store}

	t.Run("it records new event on POST", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodPost, "/", getJsonBufferFromEvent(t, event))
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertCallsNumber(t, len(store.registerCalls), 1)
		assertEvent(t, store.registerCalls[0], event)
	})
}

// Mocking
type StubEventStore struct {
	events        []Event
	registerCalls []Event
}

func (s *StubEventStore) GetEventById(id int) Event {
	event := Event{}

	for i, value := range s.events {
		if value.Id == id {
			event = s.events[i]
			break
		}
	}

	return event
}

func (s *StubEventStore) RegisterNewEvent(event Event) {
	s.registerCalls = append(s.registerCalls, event)
}

// Some helpers
func newGetIdRequest(id int) *http.Request {
	url := fmt.Sprintf("/%s", strconv.Itoa(id))
	request, _ := http.NewRequest(http.MethodGet, url, nil)
	return request
}

func getEventFromResponse(t *testing.T, body io.Reader) (event Event) {
	t.Helper()

	err := json.NewDecoder(body).Decode(&event)

	if err != nil {
		t.Fatalf("Unable to parse response from server %q into an Event, '%v'", body, err)
	}

	return
}

func assertStatus(t *testing.T, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("got status %d want %d", got, want)
	}
}

func assertContentType(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()

	responseContentType := response.Result().Header.Get("content-type")
	if responseContentType != want {
		t.Errorf("response did not have content-type of %v, got %v", want, responseContentType)
	}
}

func assertEvent(t *testing.T, got, want Event) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func getJsonBufferFromEvent(t *testing.T, event Event) *bytes.Buffer {
	t.Helper()

	jsonEvent, err := json.Marshal(event)

	if err != nil {
		t.Fatalf("did not get a json")
	}

	jsonBuffer := bytes.NewBuffer(jsonEvent)

	return jsonBuffer
}

func assertCallsNumber(t *testing.T, got, want int) {
	if got != want {
		t.Fatalf("got %d calls to RegisterNewEvent, want %d", got, want)
	}
}
