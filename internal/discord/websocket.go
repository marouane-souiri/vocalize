package discord

import (
	"fmt"
	"log"
	"sync"

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
}

type wsManagerImpl struct {
	mu   sync.Mutex
	once sync.Once

	url         string
	conn        *websocket.Conn
	sendChan    chan []byte
	receiveChan chan []byte
	errorChan   chan error
	closeChan   chan struct{}
	closed      bool
}

func newWSManager(url string) wsManager {
	return &wsManagerImpl{
		url:         url,
		sendChan:    make(chan []byte, 100),
		receiveChan: make(chan []byte, 100),
		errorChan:   make(chan error, 10),
		closeChan:   make(chan struct{}),
	}
}

func (m *wsManagerImpl) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, _, err := websocket.DefaultDialer.Dial(m.url, nil)
	if err != nil {
		return fmt.Errorf("websocket dial error: %w", err)
	}
	m.conn = c

	go m.readPump()
	go m.writePump()

	return nil
}

func (m *wsManagerImpl) Reconnect(url string) error {
	if url == "" {
		url = m.url
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.closed {
		close(m.closeChan)

		if m.conn != nil {
			_ = m.conn.Close()
			m.conn = nil
		}
	}

	m.closed = false
	m.url = url
	m.sendChan = make(chan []byte, 100)
	m.receiveChan = make(chan []byte, 100)
	m.errorChan = make(chan error, 10)
	m.closeChan = make(chan struct{})
	m.once = sync.Once{}

	c, _, err := websocket.DefaultDialer.Dial(m.url, nil)
	if err != nil {
		return fmt.Errorf("reconnect dial error: %w", err)
	}
	m.conn = c

	go m.readPump()
	go m.writePump()

	return nil
}

func (m *wsManagerImpl) Send(data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn == nil {
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

func (m *wsManagerImpl) Close() error {
	var err error

	m.once.Do(func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		if m.closed {
			return
		}
		m.closed = true

		close(m.closeChan)

		if m.conn != nil {
			err = m.conn.Close()
			m.conn = nil
		}

		close(m.sendChan)
		close(m.receiveChan)
		close(m.errorChan)
	})

	return err
}

func (m *wsManagerImpl) readPump() {
	for {
		select {
		case <-m.closeChan:
			return
		default:
			m.mu.Lock()
			conn := m.conn
			m.mu.Unlock()

			if conn == nil {
				select {
				case m.errorChan <- wsErrNotConnected:
				default:
				}
				return
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				select {
				case m.errorChan <- fmt.Errorf("read error: %w", err):
				default:
				}
				return
			}

			select {
			case m.receiveChan <- message:
			default:
				log.Println("Warning: receive channel full, dropping message")
			}
		}
	}
}

func (m *wsManagerImpl) writePump() {
	for {
		select {
		case <-m.closeChan:
			return
		case message, ok := <-m.sendChan:
			if !ok {
				return
			}

			m.mu.Lock()
			conn := m.conn
			m.mu.Unlock()

			if conn == nil {
				select {
				case m.errorChan <- wsErrNotConnected:
				default:
				}
				return
			}

			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				select {
				case m.errorChan <- fmt.Errorf("write error: %w", err):
				default:
				}
				return
			}
		}
	}
}
