package websocket

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	discordGatewayURL = "wss://gateway.discord.gg/?v=10&encoding=json"
)

func TestWSManager_Connection(t *testing.T) {
	wsm := NewWSManager()
	wsm.SetUrl(discordGatewayURL)

	if err := wsm.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if !wsm.IsConnected() {
		t.Fatalf("WebSocket should be connected")
	}

	var receivedHello bool
	select {
	case msg := <-wsm.Receive():
		var payload struct {
			Op int `json:"op"`
		}
		if err := json.Unmarshal(msg, &payload); err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		if payload.Op == 10 {
			receivedHello = true
		}

	case err := <-wsm.Errors():
		t.Fatalf("Received unexpected error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for hello message")
	}

	if !receivedHello {
		t.Fatalf("Did not receive hello message")
	}

	if err := wsm.Close(); err != nil {
		t.Fatalf("Failed to close connection: %v", err)
	}
}

func TestWSManager_SendReceive(t *testing.T) {
	wsm := NewWSManager()
	wsm.SetUrl(discordGatewayURL)

	if err := wsm.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	var receivedHello bool
	select {
	case msg := <-wsm.Receive():
		var payload struct {
			Op int `json:"op"`
		}
		if err := json.Unmarshal(msg, &payload); err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		if payload.Op == 10 {
			receivedHello = true
		}

	case err := <-wsm.Errors():
		t.Fatalf("Received unexpected error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for hello message")
	}

	if !receivedHello {
		t.Fatalf("Did not receive hello message")
	}

	heartbeat := `{"op":1,"d":null}`
	wsm.Send([]byte(heartbeat))

	var receivedAck bool
	select {
	case msg := <-wsm.Receive():
		var payload struct {
			Op int `json:"op"`
		}
		if err := json.Unmarshal(msg, &payload); err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		if payload.Op == 11 {
			receivedAck = true
		}

	case err := <-wsm.Errors():
		t.Fatalf("Received unexpected error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for heartbeat ACK")
	}

	if !receivedAck {
		t.Fatalf("Did not receive heartbeat ACK")
	}

	if err := wsm.Close(); err != nil {
		t.Fatalf("Failed to close connection: %v", err)
	}
}

func TestWSManager_Reconnect(t *testing.T) {
	wsm := NewWSManager()
	wsm.SetUrl(discordGatewayURL)

	if err := wsm.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	select {
	case msg := <-wsm.Receive():
		var payload struct {
			Op int `json:"op"`
		}
		if err := json.Unmarshal(msg, &payload); err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		if payload.Op != 10 {
			t.Fatalf("Expected hello message, got op %d", payload.Op)
		}

	case err := <-wsm.Errors():
		t.Fatalf("Received unexpected error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for hello message")
	}

	if err := wsm.Reconnect(""); err != nil {
		t.Fatalf("Failed to reconnect: %v", err)
	}

	drainErrors := func() {
		for {
			select {
			case <-wsm.Errors():
			default:
				return
			}
		}
	}

	time.Sleep(1 * time.Second)
	drainErrors()

	select {
	case msg := <-wsm.Receive():
		var payload struct {
			Op int `json:"op"`
		}
		if err := json.Unmarshal(msg, &payload); err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		if payload.Op != 10 {
			t.Fatalf("Expected hello message after reconnect, got op %d", payload.Op)
		}

	case err := <-wsm.Errors():
		t.Fatalf("Received unexpected error after reconnect: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for hello message after reconnect")
	}

	if !wsm.IsConnected() {
		t.Fatalf("WebSocket should be connected after reconnection")
	}

	heartbeat := `{"op":1,"d":null}`
	wsm.Send([]byte(heartbeat))

	select {
	case msg := <-wsm.Receive():
		var payload struct {
			Op int `json:"op"`
		}
		if err := json.Unmarshal(msg, &payload); err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		if payload.Op != 11 {
			t.Fatalf("Expected heartbeat ACK, got op %d", payload.Op)
		}

	case err := <-wsm.Errors():
		t.Fatalf("Received unexpected error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for heartbeat ACK after reconnection")
	}

	if err := wsm.Close(); err != nil {
		t.Fatalf("Failed to close connection: %v", err)
	}
}

func TestWSManager_ConcurrentOperations(t *testing.T) {
	wsm := NewWSManager()
	wsm.SetUrl(discordGatewayURL)

	if err := wsm.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	select {
	case <-wsm.Receive():
	case err := <-wsm.Errors():
		t.Fatalf("Received unexpected error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for hello message")
	}

	var wg sync.WaitGroup
	var errors []error
	var errorsMutex sync.Mutex

	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			heartbeat := fmt.Sprintf(`{"op":1,"d":%d}`, id)
			wsm.Send([]byte(heartbeat))

			select {
			case msg := <-wsm.Receive():
				var payload struct {
					Op int `json:"op"`
				}
				if err := json.Unmarshal(msg, &payload); err != nil {
					errorsMutex.Lock()
					errors = append(errors, fmt.Errorf("Failed to unmarshal message: %v", err))
					errorsMutex.Unlock()
					return
				}

				if payload.Op != 11 {
					errorsMutex.Lock()
					errors = append(errors, fmt.Errorf("Expected heartbeat ACK, got op %d", payload.Op))
					errorsMutex.Unlock()
				}

			case err := <-wsm.Errors():
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("Received unexpected error: %v", err))
				errorsMutex.Unlock()
			case <-time.After(5 * time.Second):
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("Timeout waiting for heartbeat ACK"))
				errorsMutex.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if len(errors) > 0 {
		for _, err := range errors {
			t.Errorf("Concurrent operation error: %v", err)
		}
		t.Fatalf("Failed concurrent operations")
	}

	if err := wsm.Close(); err != nil {
		t.Fatalf("Failed to close connection: %v", err)
	}
}

func TestWSManager_ErrorHandling(t *testing.T) {
	wsm := NewWSManager()
	wsm.SetUrl("wss://invalid.example.com")

	err := wsm.Connect()
	if err == nil {
		t.Fatalf("Expected connection error with invalid URL")
		wsm.Close()
	}

	wsm = NewWSManager()
	wsm.SetUrl(discordGatewayURL)

	if err := wsm.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	select {
	case <-wsm.Receive():
	case err := <-wsm.Errors():
		t.Fatalf("Received unexpected error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for hello message")
	}

	wsm.Send([]byte("this is not valid JSON"))

	select {
	case <-wsm.Receive():
	case <-wsm.Errors():
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for response after sending invalid data")
	}

	if err := wsm.Close(); err != nil {
		t.Fatalf("Failed to close connection: %v", err)
	}
}

func TestWSManager_CloseAndReopen(t *testing.T) {
	wsm := NewWSManager()
	wsm.SetUrl(discordGatewayURL)

	if err := wsm.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	select {
	case <-wsm.Receive():
	case err := <-wsm.Errors():
		t.Fatalf("Received unexpected error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for hello message")
	}

	if err := wsm.Close(); err != nil {
		t.Fatalf("Failed to close connection: %v", err)
	}

	wsm = NewWSManager()
	wsm.SetUrl(discordGatewayURL)

	if err := wsm.Connect(); err != nil {
		t.Fatalf("Failed to connect after closing: %v", err)
	}

	select {
	case <-wsm.Receive():
	case err := <-wsm.Errors():
		t.Fatalf("Received unexpected error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for hello message")
	}

	if !wsm.IsConnected() {
		t.Fatalf("WebSocket should be connected after reopening")
	}

	if err := wsm.Close(); err != nil {
		t.Fatalf("Failed to close connection: %v", err)
	}
}

func TestWSManager_SendAfterClose(t *testing.T) {
	wsm := NewWSManager()
	wsm.SetUrl(discordGatewayURL)

	if err := wsm.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	select {
	case <-wsm.Receive():
	case err := <-wsm.Errors():
		t.Fatalf("Received unexpected error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for hello message")
	}

	if err := wsm.Close(); err != nil {
		t.Fatalf("Failed to close connection: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	wsm.Send([]byte(`{"op":1,"d":null}`))

	select {
	case err := <-wsm.Errors():
		if err.Error() != WsErrNotConnected.Error() &&
			!strings.Contains(err.Error(), "closed") &&
			!strings.Contains(err.Error(), "use of closed network connection") {
			t.Fatalf("Expected connection closed error, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("Timeout waiting for connection error")
	}
}
