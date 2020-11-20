package main

type PostGreStore struct{}

func (p PostGreStore) GetEventById(id int) (event Event) {
	return
}

func (p PostGreStore) GetAllEvents() (eventList []Event) {
	return
}

func (p PostGreStore) GetEventsByFlag(flag int) (eventList []Event) {
	return
}

func (p PostGreStore) RegisterNewEvents(eventList []Event) (insertedLines int) {
	return
}
