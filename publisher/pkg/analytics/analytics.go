package analytics

import "time"

type Message struct {
	Timestamp time.Time
	Subject   string
	Data      []byte
}

type Sender func(data any, subjects ...string) error
type Handler func(msg Message, sender Sender) error

type Analytics interface {
	// Handlers returns a mapping of subjects to respective handlers.
	// It is expected for the service to subscribe to these subjects and call the handler.
	Handlers() map[string]Handler
}
