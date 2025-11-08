package models

// Event represents a domain event with a callback handler
type Event struct {
	Name    string
	Handler func(data interface{}) error
}

// EventDTO is a DTO version without the callback
type EventDTO struct {
	Name string
}

// Message with channel
type Message struct {
	ID      string
	Content string
	Notify  chan string
}

// MessageDTO without channel
type MessageDTO struct {
	ID      string
	Content string
}
