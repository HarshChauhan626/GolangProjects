package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 22: PUB/SUB SYSTEM
// ==================================================
//
// PROBLEM STATEMENT:
// A Pub/Sub (Publish-Subscribe) system extends the event bus concept with
// topic-based subscriptions. Subscribers receive messages on specific topics
// via their own dedicated channels. This decouples publishers from
// subscribers and allows fan-out delivery with backpressure.
//
// KEY DIFFERENCES FROM EVENT BUS:
// - Subscribers receive messages via channels (not callbacks)
// - Each subscriber has its own buffered channel (backpressure support)
// - Topic-based routing (similar to MQTT, Redis Pub/Sub, Kafka topics)
//
// ARCHITECTURE:
//
//   Publisher → Publish("sports", msg) → ┌────────────┐
//                                        │  Pub/Sub   │ → Subscriber A (ch) [sports]
//   Publisher → Publish("news", msg) →   │  Broker    │ → Subscriber B (ch) [sports, news]
//                                        └────────────┘ → Subscriber C (ch) [news]
//

// Message represents a published message on a topic.
type Message struct {
	Topic     string
	Payload   interface{}
	Timestamp time.Time
}

// Subscriber represents a single subscriber with its own message channel.
type Subscriber struct {
	ID       string
	Topics   []string
	Messages chan Message
	quit     chan struct{}
}

// PubSub is a topic-based publish-subscribe message broker.
type PubSub struct {
	mu          sync.RWMutex
	subscribers map[string][]*Subscriber // topic -> list of subscribers
}

// NewPubSub creates a new Pub/Sub broker.
func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: make(map[string][]*Subscriber),
	}
}

// Subscribe creates a new subscriber for the given topics.
// Returns a Subscriber with a Messages channel to read from.
func (ps *PubSub) Subscribe(id string, bufferSize int, topics ...string) *Subscriber {
	sub := &Subscriber{
		ID:       id,
		Topics:   topics,
		Messages: make(chan Message, bufferSize),
		quit:     make(chan struct{}),
	}

	ps.mu.Lock()
	for _, topic := range topics {
		ps.subscribers[topic] = append(ps.subscribers[topic], sub)
	}
	ps.mu.Unlock()

	fmt.Printf("[Broker] %s subscribed to topics: %v\n", id, topics)
	return sub
}

// Publish sends a message to all subscribers of the given topic.
// Non-blocking: if a subscriber's buffer is full, the message is dropped
// for that subscriber (prevents slow consumers from blocking publishers).
func (ps *PubSub) Publish(topic string, payload interface{}) {
	ps.mu.RLock()
	subs := make([]*Subscriber, len(ps.subscribers[topic]))
	copy(subs, ps.subscribers[topic])
	ps.mu.RUnlock()

	msg := Message{
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	for _, sub := range subs {
		select {
		case sub.Messages <- msg:
			// Message delivered
		default:
			// Subscriber's buffer is full — drop message (backpressure)
			fmt.Printf("[Broker] ⚠️  Dropped message for %s (buffer full)\n", sub.ID)
		}
	}
}

// Unsubscribe removes a subscriber from all its topics and closes its channel.
func (ps *PubSub) Unsubscribe(sub *Subscriber) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, topic := range sub.Topics {
		subs := ps.subscribers[topic]
		for i, s := range subs {
			if s.ID == sub.ID {
				ps.subscribers[topic] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}

	close(sub.quit)
	close(sub.Messages)
	fmt.Printf("[Broker] %s unsubscribed from all topics.\n", sub.ID)
}

// startSubscriber launches a goroutine that processes messages for a subscriber.
func startSubscriber(sub *Subscriber, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range sub.Messages {
			fmt.Printf("  [%s] Received on '%s': %v\n",
				sub.ID, msg.Topic, msg.Payload)
			// Simulate processing
			time.Sleep(time.Duration(10+rand.Intn(30)) * time.Millisecond)
		}
		fmt.Printf("  [%s] Channel closed. Exiting.\n", sub.ID)
	}()
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("        TUTORIAL 22: PUB/SUB SYSTEM                ")
	fmt.Println("==================================================")

	broker := NewPubSub()
	var wg sync.WaitGroup

	// 1. Create subscribers with different topic interests
	fmt.Println("\n--- Setting up subscribers ---")
	sportsReader := broker.Subscribe("SportsReader", 10, "sports")
	newsReader := broker.Subscribe("NewsReader", 10, "news")
	allReader := broker.Subscribe("AllReader", 10, "sports", "news", "weather")
	weatherReader := broker.Subscribe("WeatherReader", 10, "weather")

	// 2. Start consumer goroutines for each subscriber
	startSubscriber(sportsReader, &wg)
	startSubscriber(newsReader, &wg)
	startSubscriber(allReader, &wg)
	startSubscriber(weatherReader, &wg)

	// 3. Publish messages to various topics
	fmt.Println("\n--- Publishing messages ---")

	broker.Publish("sports", "🏀 Lakers win the championship!")
	broker.Publish("news", "📰 New trade agreement signed.")
	broker.Publish("weather", "🌤️  Sunny, 25°C expected tomorrow.")
	broker.Publish("sports", "⚽ World Cup qualifiers begin.")
	broker.Publish("news", "📰 Tech company announces new product.")
	broker.Publish("weather", "🌧️  Rain expected in the afternoon.")

	// Give subscribers time to process
	time.Sleep(300 * time.Millisecond)

	// 4. Unsubscribe a reader and publish more
	fmt.Println("\n--- Unsubscribing WeatherReader ---")
	broker.Unsubscribe(weatherReader)

	fmt.Println("\n--- Publishing more weather (only AllReader should receive) ---")
	broker.Publish("weather", "❄️  Snow alert for northern regions!")

	time.Sleep(200 * time.Millisecond)

	// 5. Clean up
	fmt.Println("\n--- Shutting down ---")
	broker.Unsubscribe(sportsReader)
	broker.Unsubscribe(newsReader)
	broker.Unsubscribe(allReader)

	wg.Wait()

	fmt.Println("\nPub/Sub system demo complete!")
	fmt.Println("Tutorial 22 complete!")
}
