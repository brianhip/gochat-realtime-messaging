package services

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait is the time allowed to write a single message to the client
	writeWait = 10 * time.Second

	// pongWait is the time allowed to read the next pong message from the client
	pongWait = 60 * time.Second

	// pingPeriod is how often pings are sent; must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// sendBufferSize is how many outgoing messages can be queued per client
	// before the client is considered too slow and disconnected
	sendBufferSize = 256
)

// Client wraps a WebSocket connection with a single writer goroutine
// gorilla/websocket allows only one concurrent writer per connection, so all
// outgoing messages (replies, broadcasts, pings) are queued on the send
// channel and written by WritePump, which sets a fresh deadline per write
type Client struct {
	Username string
	conn     *websocket.Conn
	send     chan []byte
	done     chan struct{}
	once     sync.Once
}

// NewClient creates a Client for an upgraded WebSocket connection
// The caller must start WritePump in its own goroutine
func NewClient(username string, conn *websocket.Conn) *Client {
	return &Client{
		Username: username,
		conn:     conn,
		send:     make(chan []byte, sendBufferSize),
		done:     make(chan struct{}),
	}
}

// PrepareRead sets the read deadline and extends it whenever a pong arrives
// Call once before the read loop starts
func (c *Client) PrepareRead() {
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
}

// Send queues a message for delivery to the client without blocking
// If the client's queue is full, the client is too slow to keep up and is closed
func (c *Client) Send(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error encoding message for %s: %v", c.Username, err)
		return
	}

	select {
	case <-c.done:
		// Client already closed; drop the message
	case c.send <- data:
	default:
		log.Printf("Send queue full for %s, closing connection", c.Username)
		c.Close()
	}
}

// Close stops the write pump and closes the underlying connection
// It is safe to call multiple times and from multiple goroutines
func (c *Client) Close() {
	c.once.Do(func() {
		close(c.done)
		c.conn.Close()
	})
}

// WritePump is the only goroutine that writes to the connection
// It delivers queued messages and sends periodic pings until the client is closed
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case data := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Error sending message to %s: %v", c.Username, err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}
