package discord

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var wsErrNotConnected = fmt.Errorf("websocket not connected")

type wsManager interface {
	Connect() error
	Reconnect(url string) error
	Close() error
	Send(data []byte)
	Receive() <-chan []byte
	Errors() <-chan error
	IsConnected() bool
}

type wsManagerImpl struct {
	mu   sync.RWMutex
	url  string
	conn *websocket.Conn

	sendChan    chan []byte
	receiveChan chan []byte
	errorChan   chan error
	reconnect   chan string
	shutdown    chan struct{}

	isActive bool
}

func newWSManager(url string) wsManager {
	wm := &wsManagerImpl{
		url:         url,
		sendChan:    make(chan []byte, 100),
		receiveChan: make(chan []byte, 100),
		errorChan:   make(chan error, 10),
		reconnect:   make(chan string, 5),
		shutdown:    make(chan struct{}),
	}

	go wm.connectionManager()

	go wm.readPump()
	go wm.writePump()

	return wm
}

func (m *wsManagerImpl) connectionManager() {
	for {
		select {
		case <-m.shutdown:
			m.mu.Lock()
			if m.conn != nil {
				m.conn.Close()
				m.conn = nil
			}
			m.isActive = false
			m.mu.Unlock()
			return

		case newURL := <-m.reconnect:
			m.mu.Lock()

			if newURL != "" {
				m.url = newURL
			}

			if m.conn != nil {
				m.conn.Close()
				m.conn = nil
			}

			c, _, err := websocket.DefaultDialer.Dial(m.url, nil)
			if err != nil {
				m.isActive = false
				m.mu.Unlock()
				m.errorChan <- fmt.Errorf("websocket reconnect error: %w", err)
				continue
			}

			m.conn = c
			m.isActive = true
			m.mu.Unlock()
		}
	}
}

func (m *wsManagerImpl) Connect() error {
	m.mu.RLock()
	alreadyConnected := m.isActive && m.conn != nil
	m.mu.RUnlock()

	if alreadyConnected {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	c, _, err := websocket.DefaultDialer.Dial(m.url, nil)
	if err != nil {
		m.isActive = false
		return fmt.Errorf("websocket connect error: %w", err)
	}

	m.conn = c
	m.isActive = true

	return nil
}

func (m *wsManagerImpl) Reconnect(url string) error {
	select {
	case m.reconnect <- url:
		return nil
	default:
		return fmt.Errorf("reconnect channel full")
	}
}

func (m *wsManagerImpl) Close() error {
	close(m.shutdown)
	return nil
}

func (m *wsManagerImpl) Send(data []byte) {
	m.mu.RLock()
	active := m.isActive && m.conn != nil
	m.mu.RUnlock()

	if !active {
		select {
		case m.errorChan <- wsErrNotConnected:
		default:
			log.Println("Warning: error channel full, dropping wsErrNotConnected")
		}
		return
	}

	select {
	case m.sendChan <- data:
	default:
		log.Println("Warning: send channel full, dropping message")
	}
}

func (m *wsManagerImpl) Receive() <-chan []byte {
	return m.receiveChan
}

func (m *wsManagerImpl) Errors() <-chan error {
	return m.errorChan
}

func (m *wsManagerImpl) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isActive && m.conn != nil
}

func (m *wsManagerImpl) readPump() {
	for {
		select {
		case <-m.shutdown:
			return
		default:
		}

		m.mu.RLock()
		conn := m.conn
		active := m.isActive
		m.mu.RUnlock()

		if conn == nil || !active {
			if active {
				select {
				case m.errorChan <- wsErrNotConnected:
				default:
				}

				m.mu.Lock()
				m.isActive = false
				m.mu.Unlock()
			}

			select {
			case <-m.shutdown:
				return
			case <-time.After(100 * time.Millisecond):
			}
			continue
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			select {
			case m.errorChan <- fmt.Errorf("read error: %w", err):
			default:
			}

			m.mu.Lock()
			if m.conn == conn {
				m.conn = nil
				m.isActive = false
			}
			m.mu.Unlock()

			continue
		}

		select {
		case m.receiveChan <- message:
		default:
			log.Println("Warning: receive channel full, dropping message")
		}
	}
}

func (m *wsManagerImpl) writePump() {
	for {
		select {
		case <-m.shutdown:
			return
		case message, ok := <-m.sendChan:
			if !ok {
				return
			}

			m.mu.RLock()
			conn := m.conn
			active := m.isActive
			m.mu.RUnlock()

			if conn == nil || !active {
				select {
				case m.errorChan <- wsErrNotConnected:
				default:
				}
				continue
			}

			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				select {
				case m.errorChan <- fmt.Errorf("write error: %w", err):
				default:
				}

				m.mu.Lock()
				if m.conn == conn {
					m.conn = nil
					m.isActive = false
				}
				m.mu.Unlock()
			}
		}
	}
}
