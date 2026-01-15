package observer

import (
	"fmt"

	"github.com/Elissbar/go-shortener-url/internal/model"
)

type Publisher interface {
	Subscribe(Observer)
	Unsubscribe(Observer)
	Notify()
}

type Observer interface {
	Update(model.AuditRequest) error
	GetID() string
}

type Event struct {
	Subscribers map[string]Observer
	Message     model.AuditRequest
}

func NewEvent() *Event {
	return &Event{
		Subscribers: make(map[string]Observer),
	}
}

func (e *Event) Subscribe(subs []Observer) {
	for _, sub := range subs {
		subName := sub.GetID()
		if _, ok := e.Subscribers[subName]; !ok {
			e.Subscribers[subName] = sub
		}
	}
}

func (e *Event) Unsubscribe(sub Observer) {
	delete(e.Subscribers, sub.GetID())
}

func (e *Event) Notify() {
	for _, sub := range e.Subscribers {
		sub.Update(e.Message)
	}
}

func (e *Event) Update(message model.AuditRequest) {
	fmt.Println("Message: ", message)
	e.Message = message
	e.Notify()
}
