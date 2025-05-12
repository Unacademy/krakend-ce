package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Event struct {
	ID        string    `json:"id"`
	Data      string    `json:"data"`
	Event     string    `json:"event"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Create a channel to send events
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		// Log when we start sending
		startTime := time.Now()
		fmt.Printf("Server: Starting to send events at %v\n", startTime.Format(time.StampMilli))

		// Send events with increasing delays to simulate real-world behavior
		for i := 0; i < 5; i++ {
			// Intentional delay to simulate processing time
			delay := time.Duration(i+1) * 1 * time.Second
			time.Sleep(delay)

			// Create and send event
			currentTime := time.Now()
			event := Event{
				ID:        fmt.Sprintf("%d", i),
				Data:      fmt.Sprintf("Event %d with %ds delay", i, i+1),
				Event:     "test",
				Timestamp: currentTime,
			}

			data, _ := json.Marshal(event)

			fmt.Printf("Server: Sending event %d at %v (delay: %v)\n",
				i, currentTime.Format(time.StampMilli), delay)

			fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", event.ID, event.Event, data)
			flusher.Flush()
		}

		fmt.Printf("Server: Finished sending all events. Total time: %v\n",
			time.Since(startTime))
	})

	fmt.Println("SSE test backend running on :8081")
	fmt.Println("This server sends events with increasing delays to test streaming behavior")
	http.ListenAndServe(":8081", nil)
}
