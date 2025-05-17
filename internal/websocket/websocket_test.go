package websocket

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

const TestServer = "wss://ws.postman-echo.com/raw"

func TestConnection(t *testing.T) {
	wsm := NewWSManager()
	wsm.SetUrl(TestServer)
	if err := wsm.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	if !wsm.IsConnected() {
		t.Fatalf("Should be connected")
	}
	if err := wsm.Close(); err != nil {
		t.Fatalf("Close conn failed: %v", err)
	}
}

func TestSendReceive(t *testing.T) {
	wsm := NewWSManager()
	wsm.SetUrl(TestServer)
	if err := wsm.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	testMsg := "test message"
	wsm.Send([]byte(testMsg))

	select {
	case msg := <-wsm.Receive():
		if string(msg) != testMsg {
			t.Fatalf("Expected to receive '%s', got '%s'", testMsg, string(msg))
		}
	case err := <-wsm.Errors():
		t.Fatalf("Received error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatalf("Timeout waiting for response")
	}

	wsm.Close()
}

func TestReconnect(t *testing.T) {
	wsm := NewWSManager()
	wsm.SetUrl(TestServer)
	if err := wsm.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	if err := wsm.Reconnect(""); err != nil {
		t.Fatalf("Reconnect failed: %v", err)
	}
	if !wsm.IsConnected() {
		t.Fatalf("Should be connected after reconnect")
	}
	wsm.Close()
}

func TestErrors(t *testing.T) {
	wsm := NewWSManager()
	wsm.SetUrl("ws://test.invalid.com")
	if err := wsm.Connect(); err == nil {
		t.Fatalf("Should fail to connect to invalid URL")
		wsm.Close()
	}
	wsm = NewWSManager()
	wsm.Send([]byte("test"))
	select {
	case err := <-wsm.Errors():
		if err != WsErrNotConnected {
			t.Fatalf("Expected not connected error, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("Should get error when sending while not connected")
	}
}

func TestConcurrent(t *testing.T) {
	wsm := NewWSManager()
	wsm.SetUrl(TestServer)
	if err := wsm.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	var wg sync.WaitGroup
	count := 5

	for i := range count {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msg := fmt.Sprintf("message %d", id)
			wsm.Send([]byte(msg))
		}(i)
	}

	wg.Wait()

	received := 0
	timeout := time.After(3 * time.Second)

	for received < count {
		select {
		case <-wsm.Receive():
			received++
		case err := <-wsm.Errors():
			t.Logf("Received error: %v", err)
		case <-timeout:
			break
		}
	}

	wsm.Close()
}
