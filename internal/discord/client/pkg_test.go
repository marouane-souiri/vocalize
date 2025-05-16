package client

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/marouane-souiri/vocalize/internal/discord/cache"
	"github.com/marouane-souiri/vocalize/internal/discord/websocket"
	"github.com/marouane-souiri/vocalize/internal/workerpool"
)

const testToken = "test_token"

func getClientOption() (*CLientOptions, *mockWSManager) {
	mockWs := newMockWSManager()
	wp := workerpool.NewWorkerPool(10, 5)
	cm := cache.NewDiscordCacheManager()

	return &CLientOptions{
		Ws:    mockWs,
		Wp:    wp,
		Cm:    cm,
		Token: testToken,
	}, mockWs
}

func TestClient_BasicFunctionality(t *testing.T) {
	opt, mockWs := getClientOption()

	client, err := NewClient(opt)
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
	for range 5 {
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
	opt, mockWs := getClientOption()

	client, err := NewClient(opt)
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
	opt, mockWs := getClientOption()

	client, err := NewClient(opt)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Start()
	if err != nil {
		t.Fatalf("Failed to start client: %v", err)
	}

	mockWs.injectReceiveMessage([]byte(`{"op":10,"d":{"heartbeat_interval":300}}`))

	var heartbeatCount int
	for range 20 {
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
	opt, mockWs := getClientOption()

	client, err := NewClient(opt)
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
	mockWs.injectError(websocket.WsErrNotConnected)

	time.Sleep(2 * time.Second)

	if mockWs.reconnectCount == 0 {
		t.Fatalf("Client did not attempt to reconnect")
	}

	mockWs.injectReceiveMessage([]byte(`{"op":10,"d":{"heartbeat_interval":1000}}`))

	var resumeSent bool
	for range 10 {
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
	opt, mockWs := getClientOption()

	client, err := NewClient(opt)
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

func TestClient_MassivePingLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high load test in short mode")
	}

	opt, mockWs := getClientOption()

	client, err := NewClient(opt)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	var (
		msgProcessed int32
		mutex        sync.Mutex
		startTimes   = make(map[int]time.Time)
		endTimes     = make(map[int]time.Time)
	)

	client.On("MESSAGE_CREATE", func(data json.RawMessage) {
		var message struct {
			ID      string `json:"id"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(data, &message); err != nil {
			t.Logf("Error unmarshaling message: %v", err)
			return
		}

		if message.Content == "ping" {
			msgID, _ := strconv.Atoi(message.ID)

			mutex.Lock()
			startTimes[msgID] = time.Now()
			mutex.Unlock()

			time.Sleep(500 * time.Millisecond)

			t.Logf("PONG for message %d (worker %d)", msgID, atomic.LoadInt32(&msgProcessed))

			mutex.Lock()
			endTimes[msgID] = time.Now()
			mutex.Unlock()

			atomic.AddInt32(&msgProcessed, 1)
		}
	})

	err = client.Start()
	if err != nil {
		t.Fatalf("Failed to start client: %v", err)
	}

	mockWs.injectReceiveMessage([]byte(`{"op":10,"d":{"heartbeat_interval":1000}}`))
	mockWs.injectReceiveMessage([]byte(`{"op":0,"t":"READY","s":1,"d":{"session_id":"test_session"}}`))

	const messageCount = 100
	startTime := time.Now()

	t.Logf("Starting massive ping load test with %d messages", messageCount)

	for i := range messageCount {
		msgID := i + 100
		payload := fmt.Sprintf(`{"op":0,"t":"MESSAGE_CREATE","s":%d,"d":{"id":"%d","content":"ping"}}`, i+2, msgID)
		mockWs.injectReceiveMessage([]byte(payload))

		if i%10 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	timeout := time.After(30 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			if atomic.LoadInt32(&msgProcessed) >= messageCount {
				done <- true
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-done:
	case <-timeout:
		t.Logf("Timeout waiting for all messages to be processed. Got %d/%d",
			atomic.LoadInt32(&msgProcessed), messageCount)
	}

	totalDuration := time.Since(startTime)
	t.Logf("Processed %d messages in %v", atomic.LoadInt32(&msgProcessed), totalDuration)

	var maxConcurrent int
	var totalOverlaps int

	type event struct {
		time  time.Time
		start bool
		msgID int
	}

	events := make([]event, 0, messageCount*2)

	mutex.Lock()
	for id, startTime := range startTimes {
		if endTime, ok := endTimes[id]; ok {
			events = append(events, event{time: startTime, start: true, msgID: id})
			events = append(events, event{time: endTime, start: false, msgID: id})
		}
	}
	mutex.Unlock()

	sort.Slice(events, func(i, j int) bool {
		return events[i].time.Before(events[j].time)
	})

	concurrent := 0
	concurrentIDs := make(map[int]bool)
	concurrentPairs := make(map[string]bool)

	for _, e := range events {
		if e.start {
			concurrent++
			concurrentIDs[e.msgID] = true

			for id := range concurrentIDs {
				if id != e.msgID {
					pair := fmt.Sprintf("%d,%d", min(id, e.msgID), max(id, e.msgID))
					concurrentPairs[pair] = true
				}
			}

			if concurrent > maxConcurrent {
				maxConcurrent = concurrent
			}
		} else {
			concurrent--
			delete(concurrentIDs, e.msgID)
		}
	}

	totalOverlaps = len(concurrentPairs)

	t.Logf("Maximum concurrent executions: %d", maxConcurrent)
	t.Logf("Total overlapping executions: %d", totalOverlaps)

	var totalProcessingTime time.Duration
	var processedCount int

	mutex.Lock()
	for id, startTime := range startTimes {
		if endTime, ok := endTimes[id]; ok {
			totalProcessingTime += endTime.Sub(startTime)
			processedCount++
		}
	}
	mutex.Unlock()

	if processedCount > 0 {
		avgProcessingTime := totalProcessingTime / time.Duration(processedCount)
		t.Logf("Average processing time: %v", avgProcessingTime)
		t.Logf("Theoretical sequential time: %v", avgProcessingTime*time.Duration(processedCount))
		t.Logf("Actual parallel time: %v (%.2fx speedup)", totalDuration,
			float64(avgProcessingTime*time.Duration(processedCount))/float64(totalDuration))
	}

	if maxConcurrent < 2 {
		t.Errorf("Failed to achieve parallelism, max concurrent executions: %d", maxConcurrent)
	}

	err = client.Stop()
	if err != nil {
		t.Fatalf("Failed to stop client: %v", err)
	}
}

func TestClient_CachesGuildsOnReady(t *testing.T) {
	opt, mockWs := getClientOption()

	client, err := NewClient(opt)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Start()
	if err != nil {
		t.Fatalf("Failed to start client: %v", err)
	}

	mockWs.injectReceiveMessage([]byte(`{"op":10,"d":{"heartbeat_interval":1000}}`))
	mockWs.injectReceiveMessage([]byte(`{
		"op": 0,
		"t": "READY",
		"s": 1,
		"d": {
			"guilds": [
		        {"id": "guild1", "name": "Guild One"},
		        {"id": "guild2", "name": "Guild Two"}
			],
			"session_id": "test_session",
			"resume_gateway_url": "wss://test.gateway"
		}
	}`))

	drainChannel(mockWs.sendCh)

	time.Sleep(1 * time.Second)

	cachedGuilds := client.GetGuilds()
	if len(cachedGuilds) != 2 {
		t.Fatalf("Expected 2 guilds to be cached, but found %d", len(cachedGuilds))
	}

	for _, guild := range cachedGuilds {
		t.Log(guild, "\n")
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

func (m *mockWSManager) SetUrl(url string) {
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
