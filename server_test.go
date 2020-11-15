package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestGetByIdRequest(t *testing.T) {
	event1 := Event{
		Id:        1,
		Timestamp: time.Date(2020, time.November, 15, 23, 51, 8, 84496744, time.UTC),
		Flags:     []int{7, 5},
		Data:      `{"location": "FR"}`,
	}
	event2 := Event{
		Id:        2,
		Timestamp: time.Date(2020, time.June, 7, 7, 52, 45, 575963, time.UTC),
		Flags:     []int{15, 2, 8},
		Data:      "{}",
	}

	store := StubEventStore{[]Event{event1, event2}}
	server := Server{&store}

	t.Run("Get request should return event with id 1", func(t *testing.T) {
		request := newGetRequest("/api/get/1")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertEvent(t, getEventFromResponse(t, response.Body), event1)
	})

	t.Run("Get request should return event with id 2", func(t *testing.T) {
		request := newGetRequest("/api/get/2")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertEvent(t, getEventFromResponse(t, response.Body), event2)
	})

	t.Run("returned json has lowercase key", func(t *testing.T) {
		eventAsBytes := []byte(`{"id":1,"timestamp":"2020-11-15T23:51:08.084496744Z","flags":[7,5],"data":"{\"location\": \"FR\"}"}
`)

		request := newGetRequest("/api/get/1")
		response, buffer := getRecorderWithBuffer()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertRequestBody(t, buffer, eventAsBytes)
	})

	t.Run("Get request should return status code 404 when no event with given id exists", func(t *testing.T) {
		request := newGetRequest("/api/get/5")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("Get request should return code status 422 when id given is not a number", func(t *testing.T) {
		request := newGetRequest("/api/get/aaa")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnprocessableEntity)
	})
}

// Test doubles
type StubEventStore struct {
	events []Event
}

func (s *StubEventStore) GetEventById(id int) (event Event) {
	for i, value := range s.events {
		if value.Id == id {
			event = s.events[i]
			break
		}
	}

	return
}

// Helpers: check assertions
func assertStatus(t *testing.T, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("incorrect status code: got %d, want %d", got, want)
	}
}

func assertContentType(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()

	got := response.Result().Header.Get("content-type")
	if got != want {
		t.Fatalf("incorrect content-type: got %s, want %s", got, want)
	}
}

func assertEvent(t *testing.T, got, want Event) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("event are differents: got %v, want %v", got, want)
	}
}

func assertRequestBody(t *testing.T, buffer *bytes.Buffer, want []byte) {
	got, _ := ioutil.ReadAll(buffer)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("incorrect request body:\ngot %v\n%s\nwant %v\n%s\n", got, string(got), want, string(want))
	}
}

// Helpers: tools
func getEventFromResponse(t *testing.T, body io.Reader) (event Event) {
	t.Helper()

	err := json.NewDecoder(body).Decode(&event)

	if err != nil {
		t.Fatalf("unable to parse response from server %q into an Event: %v", body, err)
	}

	return
}

func newGetRequest(target string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, target, nil)
	return request
}

func getRecorderWithBuffer() (*httptest.ResponseRecorder, *bytes.Buffer) {
	response := httptest.NewRecorder()
	buffer := &bytes.Buffer{}
	response.Body = buffer

	return response, buffer
}
