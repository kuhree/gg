package events

import (
	"fmt"
)

type GameEvent int

type EventCallback[T any] func(T)

type EventManager[T any] struct {
    listeners map[GameEvent][]EventCallback[T]
}

func NewEventManager[T any]() *EventManager[T] {
    return &EventManager[T]{
        listeners: make(map[GameEvent][]EventCallback[T]),
    }
}

func (es *EventManager[T]) Subscribe(event GameEvent, callback EventCallback[T]) {
    es.listeners[event] = append(es.listeners[event], callback)
}

func (es *EventManager[T]) UnsubscribeAll(event GameEvent) error {
    if _, ok := es.listeners[event]; !ok {
        return fmt.Errorf("event not found")
    }

    delete(es.listeners, event)
    return nil
}

func (es *EventManager[T]) Emit(event GameEvent, game T) {
    for _, callback := range es.listeners[event] {
        callback(game)
    }
}
