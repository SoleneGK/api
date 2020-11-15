package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
			Data:      `{"Age": 32}`,
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

	t.Run("returned json has keys in full lowercase", func(t *testing.T) {
		eventAsBytes := []byte(`{"id":1,"timestamp":"2020-11-11T16:04:55+01:00","author":"Solène","data":"{\"Age\": 32}"}
`)
		event := Event{}
		_ = json.Unmarshal(eventAsBytes, &event)

		request := newGetIdRequest(1)
		response := httptest.NewRecorder()
		buffer := &bytes.Buffer{}
		response.Body = buffer

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertJsonBody(t, buffer, eventAsBytes)

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

func TestCreateRequest(t *testing.T) {
	event := []Event{
		Event{
			Timestamp: time.Unix(1605107095, 0),
			Data:      "{}",
		},
		Event{
			Timestamp: time.Unix(1605107095, 0),
			Author:    "Lauren",
			Data:      "",
		},
		Event{
			Data: `{"ip":"127.0.0.1"}`,
		},
	}

	store := StubEventStore{
		events:        event,
		registerCalls: nil,
	}
	server := &Server{&store}

	t.Run("it records new event on POST", func(t *testing.T) {
		request := newPostRequest(store.events[0])
		response := httptest.NewRecorder()

		initialCallsNumber := len(store.registerCalls)
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusCreated)
		assertCallsNumber(t, len(store.registerCalls), initialCallsNumber+1)
		assertEvent(t, store.registerCalls[initialCallsNumber], store.events[0])
	})

	t.Run("the event must not be registered if data is empty", func(t *testing.T) {
		request := newPostRequest(store.events[1])
		response := httptest.NewRecorder()

		initialCallsNumber := len(store.registerCalls)
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnprocessableEntity)
		assertCallsNumber(t, len(store.registerCalls), initialCallsNumber)
	})

	t.Run("the event must not be registered if timestamp and author are empty", func(t *testing.T) {
		request := newPostRequest(store.events[2])
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnprocessableEntity)
	})
}

func TestDeleteEvent(t *testing.T) {
	events := []Event{
		Event{
			Id:        1,
			Timestamp: time.Unix(1605107095, 0),
			Author:    "Solène",
			Data:      `{"Age": 32}`,
		},
	}

	store := StubEventStore{events: events}
	server := &Server{&store}

	t.Run("DELETE request should remove event with given id", func(t *testing.T) {
		request := newDeleteRequest("/1")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNoContent)
		assertCallsNumber(t, len(store.registerCalls), 1)

		deletedEvent := store.GetEventById(1)
		if !reflect.DeepEqual(deletedEvent, Event{}) {
			t.Errorf("the event has not been deleted")
		}
	})

	t.Run("DELETE request return an error when id given is not a number", func(t *testing.T) {
		request := newDeleteRequest("/aaa")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnprocessableEntity)
	})
}

func TestUnauthorizedMethods(t *testing.T) {
	server := &Server{&StubEventStore{}}

	methods := []string{
		http.MethodHead,
		http.MethodPut,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}

	for _, method := range methods {
		t.Run(method+" request should return StatusMethodNotAllowed", func(t *testing.T) {
			request, _ := http.NewRequest(method, "/", nil)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			assertStatus(t, response.Code, http.StatusMethodNotAllowed)
		})
	}
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

func (s *StubEventStore) RemoveEvent(id int) {
	s.registerCalls = append(s.registerCalls, Event{})
	s.removeEventFromSlice(id)

}

func (s *StubEventStore) removeEventFromSlice(id int) {
	newEvents := []Event{}

	for _, event := range s.events {
		if event.Id != id {
			newEvents = append(newEvents, event)
		}
	}

	s.events = newEvents
}

// Some helpers
func newGetIdRequest(id int) *http.Request {
	url := fmt.Sprintf("/%s", strconv.Itoa(id))
	request, _ := http.NewRequest(http.MethodGet, url, nil)
	return request
}

func newPostRequest(event Event) *http.Request {
	request, _ := http.NewRequest(http.MethodPost, "/", getJsonBufferFromEvent(event))
	return request
}

func newDeleteRequest(url string) *http.Request {
	request, _ := http.NewRequest(http.MethodDelete, url, nil)
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

func getJsonBufferFromEvent(event Event) *bytes.Buffer {
	jsonEvent, _ := json.Marshal(event)
	jsonBuffer := bytes.NewBuffer(jsonEvent)

	return jsonBuffer
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

func assertCallsNumber(t *testing.T, got, want int) {
	if got != want {
		t.Fatalf("got %d calls, want %d", got, want)
	}
}

func assertJsonBody(t *testing.T, buffer *bytes.Buffer, want []byte) {
	got, _ := ioutil.ReadAll(buffer)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
