package main

import (
	"fmt"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 21: EVENT BUS
// ==================================================
//
// PROBLEM STATEMENT:
// An Event Bus implements the publish-subscribe pattern within a single
// process. Components can subscribe to events by name and get notified
// when those events are published, without knowing about each other.
// This enables loose coupling between components.
//
// KEY CONCEPTS:
// - Subscribers register handler functions for specific event names
// - Publishers emit events by name with associated data
// - The bus routes events to all matching subscribers
// - Thread-safe: multiple goroutines can publish/subscribe concurrently
//
// ARCHITECTURE:
//
//   Publisher A ──publish("user.created")──▶ ┌───────────┐
//   Publisher B ──publish("order.placed")──▶ │ Event Bus  │ ──▶ Handler 1 (user.created)
//   Publisher C ──publish("user.created")──▶ │ (routing)  │ ──▶ Handler 2 (user.created)
//                                            └───────────┘ ──▶ Handler 3 (order.placed)
//

// Event represents something that happened in the system.
type Event struct {
	Name      string
	Data      interface{}
	Timestamp time.Time
}

// EventHandler is a function that processes an event.
type EventHandler func(Event)

// Subscription represents a single subscriber.
type Subscription struct {
	ID      int
	Event   string
	Handler EventHandler
}

// EventBus manages event subscriptions and dispatching.
type EventBus struct {
	mu            sync.RWMutex
	subscribers   map[string][]Subscription
	nextID        int
}

// NewEventBus creates a new event bus.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]Subscription),
	}
}

// Subscribe registers a handler for a specific event name.
// Returns a subscription ID that can be used to unsubscribe.
func (eb *EventBus) Subscribe(eventName string, handler EventHandler) int {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.nextID++
	sub := Subscription{
		ID:      eb.nextID,
		Event:   eventName,
		Handler: handler,
	}

	eb.subscribers[eventName] = append(eb.subscribers[eventName], sub)
	return sub.ID
}

// Unsubscribe removes a subscription by its ID.
func (eb *EventBus) Unsubscribe(subID int) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for eventName, subs := range eb.subscribers {
		for i, sub := range subs {
			if sub.ID == subID {
				// Remove the subscription by swapping with last and truncating
				eb.subscribers[eventName] = append(subs[:i], subs[i+1:]...)
				return
			}
		}
	}
}

// Publish sends an event to all subscribers of that event name.
// Handlers are called synchronously in the order they subscribed.
func (eb *EventBus) Publish(eventName string, data interface{}) {
	eb.mu.RLock()
	subs := make([]Subscription, len(eb.subscribers[eventName]))
	copy(subs, eb.subscribers[eventName])
	eb.mu.RUnlock()

	event := Event{
		Name:      eventName,
		Data:      data,
		Timestamp: time.Now(),
	}

	for _, sub := range subs {
		sub.Handler(event)
	}
}

// PublishAsync sends an event to all subscribers, each in its own goroutine.
func (eb *EventBus) PublishAsync(eventName string, data interface{}, wg *sync.WaitGroup) {
	eb.mu.RLock()
	subs := make([]Subscription, len(eb.subscribers[eventName]))
	copy(subs, eb.subscribers[eventName])
	eb.mu.RUnlock()

	event := Event{
		Name:      eventName,
		Data:      data,
		Timestamp: time.Now(),
	}

	for _, sub := range subs {
		wg.Add(1)
		go func(s Subscription) {
			defer wg.Done()
			s.Handler(event)
		}(sub)
	}
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("          TUTORIAL 21: EVENT BUS                   ")
	fmt.Println("==================================================")

	bus := NewEventBus()

	// --- Subscribe handlers for different events ---

	// User service listens for user events
	bus.Subscribe("user.created", func(e Event) {
		fmt.Printf("  [UserService] New user created: %v\n", e.Data)
	})

	// Email service listens for user events to send welcome emails
	bus.Subscribe("user.created", func(e Event) {
		fmt.Printf("  [EmailService] Sending welcome email to: %v\n", e.Data)
	})

	// Analytics service listens for all events
	bus.Subscribe("user.created", func(e Event) {
		fmt.Printf("  [Analytics] Tracking event '%s' at %s\n",
			e.Name, e.Timestamp.Format("15:04:05.000"))
	})

	// Order service
	bus.Subscribe("order.placed", func(e Event) {
		fmt.Printf("  [OrderService] Processing order: %v\n", e.Data)
	})

	// Inventory service
	inventorySubID := bus.Subscribe("order.placed", func(e Event) {
		fmt.Printf("  [Inventory] Reserving stock for order: %v\n", e.Data)
	})

	// Notification service
	bus.Subscribe("order.placed", func(e Event) {
		fmt.Printf("  [Notifications] Notifying customer about order: %v\n", e.Data)
	})

	// --- Publish events ---
	fmt.Println("\n--- Publishing 'user.created' event ---")
	bus.Publish("user.created", map[string]string{
		"name":  "Alice",
		"email": "alice@example.com",
	})

	fmt.Println("\n--- Publishing 'order.placed' event ---")
	bus.Publish("order.placed", map[string]interface{}{
		"orderID": 12345,
		"items":   3,
		"total":   99.99,
	})

	// --- Demonstrate unsubscribe ---
	fmt.Println("\n--- Unsubscribing Inventory from 'order.placed' ---")
	bus.Unsubscribe(inventorySubID)

	fmt.Println("\n--- Publishing another 'order.placed' event ---")
	bus.Publish("order.placed", map[string]interface{}{
		"orderID": 12346,
		"items":   1,
		"total":   29.99,
	})

	// --- Demonstrate async publishing ---
	fmt.Println("\n--- Async Publishing 'user.created' event ---")
	var wg sync.WaitGroup
	bus.PublishAsync("user.created", map[string]string{
		"name":  "Bob",
		"email": "bob@example.com",
	}, &wg)
	wg.Wait()

	fmt.Println("\nEvent bus demo complete!")
	fmt.Println("Tutorial 21 complete!")
}
