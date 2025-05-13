package discord

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

const testToken = "test_token"

func TestClient_BasicFunctionality(t *testing.T) {
	mockWs := newMockWSManager()
	client, err := NewClient(mockWs, testToken)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Start()
	if err != nil {
		t.Fatalf("Failed to start client: %v", err)
	}

	mockWs.injectReceiveMessage([]byte(`{"op":10,"d":{"heartbeat_interval":1000}}`))

	time.Sleep(200 * time.Millisecond)

	var identifySent bool
	for i := 0; i < 5; i++ {
		select {
		case msg := <-mockWs.sendCh:
			var payload Payload
			if err := json.Unmarshal(msg, &payload); err == nil && payload.Op == opIdentify {
				identifySent = true
			}
		default:
		}
	}

	if !identifySent {
		t.Fatalf("Client did not send identify after hello")
	}

	err = client.Stop()
	if err != nil {
		t.Fatalf("Failed to stop client: %v", err)
	}
}

func TestClient_EventHandling(t *testing.T) {
	mockWs := newMockWSManager()
	client, err := NewClient(mockWs, testToken)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	var eventCount int
	var wg sync.WaitGroup
	wg.Add(1)

	client.On("MESSAGE_CREATE", func(data json.RawMessage) {
		eventCount++
		wg.Done()
	})

	err = client.Start()
	if err != nil {
		t.Fatalf("Failed to start client: %v", err)
	}

	mockWs.injectReceiveMessage([]byte(`{"op":10,"d":{"heartbeat_interval":1000}}`))
	mockWs.injectReceiveMessage([]byte(`{"op":0,"t":"READY","s":1,"d":{"session_id":"test_session"}}`))
	mockWs.injectReceiveMessage([]byte(`{"op":0,"t":"MESSAGE_CREATE","s":2,"d":{"content":"test message"}}`))

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("Event handler was not called")
	}

	if eventCount != 1 {
		t.Fatalf("Expected 1 event, got %d", eventCount)
	}

	err = client.Stop()
	if err != nil {
		t.Fatalf("Failed to stop client: %v", err)
	}
}

func TestClient_Heartbeat(t *testing.T) {
	mockWs := newMockWSManager()
	client, err := NewClient(mockWs, testToken)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Start()
	if err != nil {
		t.Fatalf("Failed to start client: %v", err)
	}

	mockWs.injectReceiveMessage([]byte(`{"op":10,"d":{"heartbeat_interval":300}}`))

	var heartbeatCount int
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)

		select {
		case msg := <-mockWs.sendCh:
			var payload Payload
			if err := json.Unmarshal(msg, &payload); err == nil && payload.Op == opHeartbeat {
				heartbeatCount++
			}
		default:
		}
	}

	if heartbeatCount < 1 {
		t.Fatalf("Client did not send heartbeats")
	}

	err = client.Stop()
	if err != nil {
		t.Fatalf("Failed to stop client: %v", err)
	}
}

func TestClient_Reconnect(t *testing.T) {
	mockWs := newMockWSManager()
	client, err := NewClient(mockWs, testToken)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Start()
	if err != nil {
		t.Fatalf("Failed to start client: %v", err)
	}

	mockWs.injectReceiveMessage([]byte(`{"op":10,"d":{"heartbeat_interval":1000}}`))
	mockWs.injectReceiveMessage([]byte(`{"op":0,"t":"READY","s":1,"d":{"session_id":"test_session","resume_gateway_url":"wss://test.gateway"}}`))

	drainChannel(mockWs.sendCh)

	mockWs.reconnectCount = 0
	mockWs.injectError(wsErrNotConnected)

	time.Sleep(2 * time.Second)

	if mockWs.reconnectCount == 0 {
		t.Fatalf("Client did not attempt to reconnect")
	}

	mockWs.injectReceiveMessage([]byte(`{"op":10,"d":{"heartbeat_interval":1000}}`))

	var resumeSent bool
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)

		select {
		case msg := <-mockWs.sendCh:
			var payload Payload
			if err := json.Unmarshal(msg, &payload); err == nil && payload.Op == opResume {
				resumeSent = true
			}
		default:
		}

		if resumeSent {
			break
		}
	}

	if !resumeSent {
		t.Fatalf("Client did not send resume after reconnect")
	}

	err = client.Stop()
	if err != nil {
		t.Fatalf("Failed to stop client: %v", err)
	}
}

func TestClient_OnceHandler(t *testing.T) {
	mockWs := newMockWSManager()
	client, err := NewClient(mockWs, testToken)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	var callCount int

	client.Once("GUILD_CREATE", func(data json.RawMessage) {
		callCount++
	})

	err = client.Start()
	if err != nil {
		t.Fatalf("Failed to start client: %v", err)
	}

	mockWs.injectReceiveMessage([]byte(`{"op":10,"d":{"heartbeat_interval":1000}}`))
	mockWs.injectReceiveMessage([]byte(`{"op":0,"t":"GUILD_CREATE","s":1,"d":{"id":"123"}}`))
	mockWs.injectReceiveMessage([]byte(`{"op":0,"t":"GUILD_CREATE","s":2,"d":{"id":"456"}}`))

	time.Sleep(300 * time.Millisecond)

	if callCount != 1 {
		t.Fatalf("Once handler called %d times, expected 1", callCount)
	}

	err = client.Stop()
	if err != nil {
		t.Fatalf("Failed to stop client: %v", err)
	}
}

func drainChannel(ch chan []byte) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

type mockWSManager struct {
	sendCh         chan []byte
	receiveCh      chan []byte
	errorCh        chan error
	reconnectCount int
	mu             sync.Mutex
}

func newMockWSManager() *mockWSManager {
	return &mockWSManager{
		sendCh:    make(chan []byte, 100),
		receiveCh: make(chan []byte, 100),
		errorCh:   make(chan error, 10),
	}
}

func (m *mockWSManager) Connect() error {
	return nil
}

func (m *mockWSManager) Reconnect(url string) error {
	m.mu.Lock()
	m.reconnectCount++
	m.mu.Unlock()
	return nil
}

func (m *mockWSManager) Close() error {
	return nil
}

func (m *mockWSManager) Send(data []byte) {
	select {
	case m.sendCh <- data:
	default:
	}
}

func (m *mockWSManager) Receive() <-chan []byte {
	return m.receiveCh
}

func (m *mockWSManager) Errors() <-chan error {
	return m.errorCh
}

func (m *mockWSManager) IsConnected() bool {
	return true
}

func (m *mockWSManager) injectReceiveMessage(data []byte) {
	m.receiveCh <- data
}

func (m *mockWSManager) injectError(err error) {
	m.errorCh <- err
}
