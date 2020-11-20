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

	store = &StubEventStore{events: []Event{event1, event2}}

	t.Run("Get request should return event with id 1", func(t *testing.T) {
		request := newGetRequest(api_url + "1")
		response := httptest.NewRecorder()

		newServer().ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertEvent(t, getEventFromResponse(t, response.Body), event1)
	})

	t.Run("Get request should return event with id 2", func(t *testing.T) {
		request := newGetRequest(api_url + "2")
		response := httptest.NewRecorder()

		newServer().ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertEvent(t, getEventFromResponse(t, response.Body), event2)
	})

	t.Run("returned json has lowercase key", func(t *testing.T) {
		eventAsBytes := []byte(`{"id":1,"timestamp":"2020-11-15T23:51:08.084496744Z","flags":[7,5],"data":"{\"location\": \"FR\"}"}
`)

		request := newGetRequest(api_url + "1")
		response, buffer := getRecorderWithBuffer()

		newServer().ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertRequestBodyBytes(t, buffer, eventAsBytes)
	})

	t.Run("Get request should return status code 404 when no event with given id exists", func(t *testing.T) {
		request := newGetRequest(api_url + "5")
		response := httptest.NewRecorder()

		newServer().ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("Get request should return code status 422 when id given is not a number", func(t *testing.T) {
		request := newGetRequest(api_url + "aaa")
		response := httptest.NewRecorder()

		newServer().ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnprocessableEntity)
	})
}

func TestGetAllRequest(t *testing.T) {
	eventList := []Event{
		Event{
			Id:        1,
			Timestamp: time.Date(2020, time.November, 15, 23, 51, 8, 84496744, time.UTC),
			Flags:     []int{7, 5},
			Data:      `{"location": "FR"}`,
		},
		{
			Id:        2,
			Timestamp: time.Date(2020, time.June, 7, 7, 52, 45, 575963, time.UTC),
			Flags:     []int{15, 2, 8},
			Data:      "{}",
		},
	}

	store = &StubEventStore{events: eventList}

	t.Run("Get request should return all events", func(t *testing.T) {
		request := newGetRequest(api_url)
		response, buffer := getRecorderWithBuffer()

		newServer().ServeHTTP(response, request)

		want, _ := json.Marshal(eventList)
		want = append(want, 10) // adding a line break

		assertStatus(t, response.Code, http.StatusOK)
		assertRequestBodyBytes(t, buffer, want)
	})
}

func TestGetByFlagRequest(t *testing.T) {
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

	store = &StubEventStore{events: []Event{event1, event2}}

	t.Run("Get request should return events with given flag", func(t *testing.T) {
		request := newGetRequest(api_url + "getFlag/8")
		response, buffer := getRecorderWithBuffer()

		newServer().ServeHTTP(response, request)

		got := []Event{}
		_ = json.NewDecoder(buffer).Decode(&got)

		assertStatus(t, response.Code, http.StatusOK)
		assertEventList(t, got, []Event{event2})
	})

	t.Run("Get request should return status code 404 when no event with given flag found", func(t *testing.T) {
		request := newGetRequest(api_url + "getFlag/9")
		response := httptest.NewRecorder()

		newServer().ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("Get request should return status code 422 when flag given is not a number", func(t *testing.T) {
		request := newGetRequest(api_url + "getFlag/auie")
		response := httptest.NewRecorder()

		newServer().ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnprocessableEntity)
	})
}

func TestPostRequest(t *testing.T) {
	event1 := Event{
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

	t.Run("Post request should call RegisterNewEvents, pass event list and return number of lines created", func(t *testing.T) {
		eventList := []Event{event1, event2}
		spy := &Spy{}
		store = &StubEventStore{eventList, spy}

		request := newPostRequest(eventList)
		response := httptest.NewRecorder()

		newServer().ServeHTTP(response, request)

		want := fmt.Sprintf("{\"%s\":%d}", lineNumberResponseKey, 2)

		assertStatus(t, response.Code, http.StatusOK)
		assertCallNumber(t, spy.callNumber, 1)
		assertEventList(t, spy.parameterGiven, eventList)
		assertRequestBodyString(t, response.Body.String(), want)
	})
}

// Test doubles
type Spy struct {
	callNumber     int
	parameterGiven []Event
}

type StubEventStore struct {
	events []Event
	spy    *Spy
	// eventListGivenAsParameter []Event
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

func (s *StubEventStore) GetAllEvents() []Event {
	return s.events
}

func (s *StubEventStore) GetEventsByFlag(flag int) (eventList []Event) {
	for _, event := range s.events {
		if contains(event.Flags, flag) {
			eventList = append(eventList, event)
		}
	}

	return
}

func contains(intSlice []int, value int) bool {
	for _, element := range intSlice {
		if element == value {
			return true
		}
	}
	return false
}

func (s *StubEventStore) RegisterNewEvents(eventList []Event) int {
	s.spy.callNumber++
	s.spy.parameterGiven = eventList
	return len(eventList)
}

// Helpers: check assertions
func assertCallNumber(t *testing.T, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("incorrect call number, got %v, want %v", got, want)
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

func assertEventList(t *testing.T, got, want []Event) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("event lists are differentt: got %v, want %v", got, want)
	}
}

func assertRequestBodyBytes(t *testing.T, buffer *bytes.Buffer, want []byte) {
	t.Helper()

	got, _ := ioutil.ReadAll(buffer)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("incorrect request body:\ngot %v\n%s\nwant %v\n%s\n", got, string(got), want, string(want))
	}
}

func assertRequestBodyString(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("ncorrect body response: got %v, want %v", got, want)
	}
}

func assertStatus(t *testing.T, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("incorrect status code: got %d, want %d", got, want)
	}
}

// Helpers: tools
func newGetRequest(target string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, target, nil)
	return request
}

func newPostRequest(eventList []Event) *http.Request {
	requestBody, _ := json.Marshal(eventList)
	requestBodyAsBytes := bytes.NewBuffer(requestBody)

	request, _ := http.NewRequest(http.MethodPost, api_url, requestBodyAsBytes)
	return request
}

func getRecorderWithBuffer() (*httptest.ResponseRecorder, *bytes.Buffer) {
	response := httptest.NewRecorder()
	buffer := &bytes.Buffer{}
	response.Body = buffer

	return response, buffer
}

func getEventFromResponse(t *testing.T, body io.Reader) (event Event) {
	t.Helper()

	err := json.NewDecoder(body).Decode(&event)

	if err != nil {
		t.Fatalf("unable to parse response from server %q into an Event: %v", body, err)
	}

	return
}
