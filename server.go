package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type EventStore interface {
	GetEventById(id int) string
}

type Server struct {
	store EventStore
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/"))
	fmt.Fprint(w, s.store.GetEventById(id))
}
