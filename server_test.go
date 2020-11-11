package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestGETEvent(t *testing.T) {
	store := StubEventStore{
		map[int]string{
			1: `{
id: 1,
author: "Solène",
}`,
			2: `{
id: 2,
author: "Camille",
}`,
		},
	}

	server := &Server{&store}

	t.Run("return event with id 1", func(t *testing.T) {
		request := newGetIdRequest(1)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		// got : le json renvoyé
		got := response.Body.String()
		// want : un json avec une id de 1
		want := `{
id: 1,
author: "Solène",
}`
		assertResponseBody(t, got, want)
	})

	t.Run("return event with id 2", func(t *testing.T) {
		request := newGetIdRequest(2)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Body.String()
		want := `{
id: 2,
author: "Camille",
}`

		assertResponseBody(t, got, want)
	})
}

// Mocking
type StubEventStore struct {
	events map[int]string
}

func (s *StubEventStore) GetEventById(id int) string {
	event := s.events[id]
	return event
}

// Some helpers
func newGetIdRequest(id int) *http.Request {
	url := fmt.Sprintf("/%s", strconv.Itoa(id))
	request, _ := http.NewRequest(http.MethodGet, url, nil)
	return request
}

func assertResponseBody(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
