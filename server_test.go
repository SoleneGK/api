package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

var server = newServer()

const (
	registerFunctionName   = "RegisterNewEvents"
	deleteByIdFunctionName = "DeleteById"
)

func TestGetByIdRequest(t *testing.T) {
	validEvent1 := Event{
		Id:        1,
		Timestamp: time.Date(2020, time.November, 15, 23, 51, 8, 84496744, time.UTC),
		Flags:     []int{7, 5},
		Data:      `{"location": "FR"}`,
	}
	validEvent2 := Event{
		Id:        2,
		Timestamp: time.Date(2020, time.June, 7, 7, 52, 45, 575963, time.UTC),
		Flags:     []int{15, 2, 8},
		Data:      "{}",
	}

	store = &StubEventStore{events: []Event{validEvent1, validEvent2}}

	t.Run("Get request should return event with id 1", func(t *testing.T) {
		request := newGetRequest(api_url + "1")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertEvent(t, getEventFromResponse(t, response.Body), validEvent1)
	})

	t.Run("Get request should return event with id 2", func(t *testing.T) {
		request := newGetRequest(api_url + "2")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertEvent(t, getEventFromResponse(t, response.Body), validEvent2)
	})

	t.Run("returned json has lowercase key", func(t *testing.T) {
		request := newGetRequest(api_url + "1")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		wantResponse := "{\"id\":1,\"timestamp\":\"2020-11-15T23:51:08.084496744Z\",\"flags\":[7,5],\"data\":\"{\\\"location\\\": \\\"FR\\\"}\"}\n"

		assertStatus(t, response.Code, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertResponseBody(t, response.Body.String(), wantResponse)
	})

	t.Run("Get request should return status code 404 when no event with given id exists", func(t *testing.T) {
		request := newGetRequest(api_url + "5")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("Get request should return code status 422 when id given is not a number", func(t *testing.T) {
		request := newGetRequest(api_url + "aaa")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

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
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		eventListAsJson, _ := json.Marshal(eventList)
		want := string(append(eventListAsJson, 10)) // adding a line break

		assertStatus(t, response.Code, http.StatusOK)
		assertResponseBody(t, response.Body.String(), want)
	})
}

func TestGetByFlagRequest(t *testing.T) {
	validEvent1 := Event{
		Id:        1,
		Timestamp: time.Date(2020, time.November, 15, 23, 51, 8, 84496744, time.UTC),
		Flags:     []int{7, 5},
		Data:      `{"location": "FR"}`,
	}
	validEvent2 := Event{
		Id:        2,
		Timestamp: time.Date(2020, time.June, 7, 7, 52, 45, 575963, time.UTC),
		Flags:     []int{15, 2, 8},
		Data:      "{}",
	}

	store = &StubEventStore{events: []Event{validEvent1, validEvent2}}

	t.Run("Get request should return events with given flag", func(t *testing.T) {
		request := newGetRequest(api_url + "getFlag/8")
		response, buffer := getRecorderWithBuffer()

		server.ServeHTTP(response, request)

		got := []Event{}
		_ = json.NewDecoder(buffer).Decode(&got)

		assertStatus(t, response.Code, http.StatusOK)
		assertEventList(t, got, []Event{validEvent2})
	})

	t.Run("Get request should return status code 404 when no event with given flag found", func(t *testing.T) {
		request := newGetRequest(api_url + "getFlag/9")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("Get request should return status code 422 when flag given is not a number", func(t *testing.T) {
		request := newGetRequest(api_url + "getFlag/auie")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnprocessableEntity)
	})
}

func TestPostRequest(t *testing.T) {
	validEvent1 := Event{
		Timestamp: time.Date(2020, time.November, 15, 23, 51, 8, 84496744, time.UTC),
		Flags:     []int{7, 5},
		Data:      `{"location": "FR"}`,
	}
	validEvent2 := Event{
		Id:        2,
		Timestamp: time.Date(2020, time.June, 7, 7, 52, 45, 575963, time.UTC),
		Flags:     []int{15, 2, 8},
		Data:      "{}",
	}
	validEvent3 := Event{
		Flags: []int{3},
		Data:  `{"Age":35}`,
	}
	invalidEvent1 := Event{
		Id:        15,
		Timestamp: time.Date(2020, time.April, 27, 19, 16, 45, 575963, time.UTC),
		Flags:     []int{9},
	}
	invalidEvent2 := Event{
		Timestamp: time.Date(2019, time.November, 21, 07, 30, 22, 658463, time.UTC),
		Data:      `{"Env":"dev"}`,
	}

	t.Run("Post request should call RegisterNewEvents, pass event list and return number of lines created", func(t *testing.T) {
		eventList := []Event{validEvent1, validEvent2}
		spy := &Spy{}
		store = &StubEventStore{eventList, spy}

		request := newPostRequest(eventList)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		want := fmt.Sprintf("{\"%s\":%d}", lineNumberResponseKey, 2)

		assertStatus(t, response.Code, http.StatusOK)
		assertCalledFunction(t, spy.calledFunction, registerFunctionName)
		assertEventList(t, spy.listGivenAsParameter, eventList)
		assertResponseBody(t, response.Body.String(), want)
	})

	t.Run("Post request should register only valid events", func(t *testing.T) {
		eventList := []Event{validEvent1, invalidEvent1, validEvent2, validEvent3, invalidEvent2}
		spy := &Spy{}
		store = &StubEventStore{eventList, spy}
		clock = MockClock{}

		request := newPostRequest(eventList)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		wantResponse := fmt.Sprintf("{\"%s\":%d}", lineNumberResponseKey, 3)

		validEvent3.Timestamp = clock.Now()
		wantlistGivenAsParameter := []Event{validEvent1, validEvent2, validEvent3}

		assertStatus(t, response.Code, http.StatusOK)
		assertCalledFunction(t, spy.calledFunction, registerFunctionName)
		assertResponseBody(t, response.Body.String(), wantResponse)
		assertEventList(t, spy.listGivenAsParameter, wantlistGivenAsParameter)
	})
}

func TestDeleteByIdRequest(t *testing.T) {
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
	event3 := Event{
		Id:    3,
		Flags: []int{3},
		Data:  `{"Age":35}`,
	}

	t.Run("Delete request should set event with given id to default values", func(t *testing.T) {
		eventList := []Event{event1, event2, event3}
		spy := &Spy{}
		store = &StubEventStore{eventList, spy}

		request := newDeleteRequest(api_url + "3")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertCalledFunction(t, spy.calledFunction, deleteByIdFunctionName)

		wantEventList := []Event{event1, event2, createNeutralEventWithId(3)}
		assertEventList(t, store.GetAllEvents(), wantEventList)

		wantResponse := fmt.Sprintf("{\"%s\":%d}", lineNumberResponseKey, 1)
		assertResponseBody(t, response.Body.String(), wantResponse)
	})

	t.Run("Delete request should return 0 lines affected when no event with given id exists", func(t *testing.T) {
		store = &StubEventStore{[]Event{}, &Spy{}}

		request := newDeleteRequest(api_url + "4")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		want := fmt.Sprintf("{\"%s\":%d}", lineNumberResponseKey, 0)

		assertStatus(t, response.Code, http.StatusOK)
		assertResponseBody(t, response.Body.String(), want)
	})

	t.Run("Delete request should return status code 422 if parameter given is not a number", func(t *testing.T) {
		store = &StubEventStore{[]Event{}, &Spy{}}

		request := newDeleteRequest(api_url + "tstnrst")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnprocessableEntity)
	})
}

// Test doubles
type Spy struct {
	calledFunction       string
	listGivenAsParameter []Event
}

type StubEventStore struct {
	events []Event
	spy    *Spy
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
	s.spy.calledFunction = registerFunctionName
	s.spy.listGivenAsParameter = eventList
	return len(eventList)
}

func (s *StubEventStore) DeleteById(id int) int {
	s.spy.calledFunction = deleteByIdFunctionName

	for i, event := range s.events {
		if event.Id == id {
			s.events[i] = createNeutralEventWithId(id)
			return 1
		}
	}

	return 0
}

type MockClock struct{}

func (m MockClock) Now() time.Time {
	return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
}

// Helpers: check assertions
func assertCalledFunction(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("incorrect function called, got %v, want %v", got, want)
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
		t.Errorf("event lists are differentt:\ngot %v\nwant %v", got, want)
	}
}

func assertResponseBody(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("incorrect body response: got %v, want %v", got, want)
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

func newDeleteRequest(target string) *http.Request {
	request, _ := http.NewRequest(http.MethodDelete, target, nil)
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

func createNeutralEventWithId(id int) Event {
	return Event{
		Id:        id,
		Timestamp: neutralTimestampValue,
		Flags:     neutralFlagsValue,
		Data:      neutralDataValue,
	}
}
