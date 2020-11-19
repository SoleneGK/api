package main

type PostGresStore struct{}

func (p PostGresStore) GetEventById(id int) (event Event) {
	return
}

func (p PostGresStore) GetAllEvents() (eventList []Event) {
	return
}

func (p PostGresStore) GetEventsByFlag(flag int) (eventList []Event) {
	return
}
