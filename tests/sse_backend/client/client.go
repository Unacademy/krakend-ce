package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Event struct {
	ID        string    `json:"id"`
	Data      string    `json:"data"`
	Event     string    `json:"event"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	// Get target URL from command line or use default
	targetURL := "http://localhost:8080/events"
	if len(os.Args) > 1 {
		targetURL = os.Args[1]
	}

	fmt.Printf("Client: Connecting to %s\n", targetURL)
	startTime := time.Now()
	fmt.Printf("Client: Starting at %v\n", startTime.Format(time.StampMilli))

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Make the SSE request
	resp, err := http.Get(targetURL)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Create a reader for the response body
	reader := bufio.NewReader(resp.Body)

	// Print headers
	fmt.Println("Response Headers:")
	for k, v := range resp.Header {
		fmt.Printf("%s: %v\n", k, v)
	}
	fmt.Println("\nEvents:")

	// Variables to track events
	var eventData string
	var eventID string
	var eventType string
	var receivedTimestamp time.Time

	// Read and print events
	for {
		select {
		case <-sigChan:
			fmt.Println("\nReceived interrupt signal, closing connection...")
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println("Connection closed by server")
					return
				}
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			// Trim the line
			line = strings.TrimSpace(line)

			// If we get an empty line, this is the end of an event
			if line == "" && eventData != "" {
				// Parse the event data
				receivedTimestamp = time.Now()

				// Print the raw event
				fmt.Print("Received at: ", receivedTimestamp.Format(time.StampMilli), "\n")
				fmt.Printf("id: %s\nevent: %s\ndata: %s\n\n", eventID, eventType, eventData)

				// Try to parse the JSON to get the server timestamp
				var event Event
				if err := json.Unmarshal([]byte(eventData), &event); err == nil {
					latency := receivedTimestamp.Sub(event.Timestamp)
					fmt.Printf("Server timestamp: %v\n", event.Timestamp.Format(time.StampMilli))
					fmt.Printf("Latency: %v\n\n", latency)
				}

				// Reset for next event
				eventData = ""
				eventID = ""
				eventType = ""
				continue
			}

			// Parse SSE fields
			if strings.HasPrefix(line, "data: ") {
				eventData = strings.TrimPrefix(line, "data: ")
			} else if strings.HasPrefix(line, "id: ") {
				eventID = strings.TrimPrefix(line, "id: ")
			} else if strings.HasPrefix(line, "event: ") {
				eventType = strings.TrimPrefix(line, "event: ")
			}
		}
	}
}
