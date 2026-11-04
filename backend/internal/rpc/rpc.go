// Package rpc hosts the FlClash core in-process, speaking the same
// length-prefixed framed JSON protocol the Flutter desktop host uses over a
// pipe: MethodCall requests in, MethodResponse answers and `message` event
// batches out.
package rpc

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"core"
)

const (
	maxFrameSize    = 64 * 1024 * 1024
	coreDialTimeout = 30 * time.Second
	callTimeout     = 30 * time.Second
	callLongTimeout = 10 * time.Minute // provider/geo updates and delay sweeps
	spillThreshold  = 1 << 20
)

// netConn abstracts the platform endpoint connection.
type netConn = net.Conn

// Event is one core event as the Dart side's CoreEventManager reads it.
type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type methodCall struct {
	ID        string          `json:"id,omitempty"`
	Method    string          `json:"method"`
	Arguments json.RawMessage `json:"arguments"`
}

type methodResponse struct {
	ID     string          `json:"id,omitempty"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details any    `json:"details"`
	} `json:"error"`
	File string `json:"file,omitempty"`
}

// Client is the host side of the core connection.
type Client struct {
	mu       sync.Mutex
	conn     net.Conn
	nextID   int
	pending  map[string]chan methodResponse
	closed   bool
	spillDir string

	// OnDisconnect fires once when the core transport drops.
	OnDisconnect func()

	eventMu     sync.Mutex
	subscribers map[int]func(Event)
	subSeq      int
	eventQueue  chan []byte
}

var ErrClosed = errors.New("core connection closed")

// Start launches the core in-process. The core dials back over the platform
// endpoint (named pipe on Windows, loopback TCP elsewhere), exactly as its
// desktop builds dial a pipe or socket, and the host speaks the framed JSON
// protocol with it.
func Start(homeDir string) (*Client, error) {
	endpoint, arg, err := listenEndpoint()
	if err != nil {
		return nil, err
	}

	client := &Client{
		pending:     map[string]chan methodResponse{},
		subscribers: map[int]func(Event){},
		spillDir:    filepath.Join(homeDir, "responses"),
		eventQueue:  make(chan []byte, 8192),
	}
	// Event frames are handled off the read loop so a subscriber may issue
	// calls back over the same connection without deadlocking it.
	go client.runEventQueue()

	// The core dials our listener once and holds the connection for its
	// lifetime; when it drops the host restarts the core.
	accepted := make(chan error, 1)
	go func() {
		conn, acceptErr := endpoint.Accept()
		endpoint.Close()
		if acceptErr != nil {
			accepted <- acceptErr
			return
		}
		accepted <- nil
		client.serve(conn)
	}()
	go coreStartServer(arg)

	// Wait until the core's dial lands so the first calls never race it. A core
	// that dies before dialing must surface as an error rather than parking the
	// host here forever with no output.
	select {
	case err := <-accepted:
		if err != nil {
			return nil, err
		}
	case <-time.After(coreDialTimeout):
		endpoint.Close()
		return nil, errors.New("core did not connect back in time")
	}
	return client, nil
}

func coreStartServer(port string) {
	defer func() { recover() }()
	core.StartServer(port)
}

func (c *Client) serve(conn net.Conn) {
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	readFrames(conn, func(data []byte) bool {
		var probe struct {
			Method string `json:"method"`
			ID     string `json:"id"`
		}
		if json.Unmarshal(data, &probe) != nil {
			return true
		}
		if probe.Method == "message" && probe.ID == "" {
			select {
			case c.eventQueue <- data:
			default:
				// queue full: drop the oldest batch and keep the newest
				select {
				case <-c.eventQueue:
				default:
				}
				select {
				case c.eventQueue <- data:
				default:
				}
			}
			return true
		}
		var response methodResponse
		if json.Unmarshal(data, &response) == nil {
			c.resolve(response.ID, response)
		}
		return true
	}, func() {
		c.mu.Lock()
		c.conn = nil
		c.closed = true
		pending := c.pending
		c.pending = map[string]chan methodResponse{}
		c.mu.Unlock()
		for _, ch := range pending {
			ch <- methodResponse{}
		}
		if c.OnDisconnect != nil {
			c.OnDisconnect()
		}
	})
}

// readFrames drains length-prefixed frames until the connection dies.
func readFrames(conn net.Conn, onFrame func([]byte) bool, onClose func()) {
	defer onClose()
	length := make([]byte, 4)
	for {
		if _, err := io.ReadFull(conn, length); err != nil {
			return
		}
		size := binary.LittleEndian.Uint32(length)
		if size > maxFrameSize {
			return
		}
		frame := make([]byte, size)
		if _, err := io.ReadFull(conn, frame); err != nil {
			return
		}
		if !onFrame(frame) {
			return
		}
	}
}

func (c *Client) resolve(id string, response methodResponse) {
	c.mu.Lock()
	ch, ok := c.pending[id]
	delete(c.pending, id)
	c.mu.Unlock()
	if ok {
		ch <- response
	}
}

func (c *Client) runEventQueue() {
	for frame := range c.eventQueue {
		c.deliverEvents(frame)
	}
}

func (c *Client) deliverEvents(frame []byte) {
	var batch struct {
		Arguments []struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		} `json:"arguments"`
	}
	if json.Unmarshal(frame, &batch) != nil {
		return
	}
	c.eventMu.Lock()
	subs := make([]func(Event), 0, len(c.subscribers))
	for _, sub := range c.subscribers {
		subs = append(subs, sub)
	}
	c.eventMu.Unlock()
	for _, item := range batch.Arguments {
		event := Event{Type: item.Type, Data: item.Data}
		for _, sub := range subs {
			sub(event)
		}
	}
}

// Subscribe registers an event listener; the id cancels via Unsubscribe.
func (c *Client) Subscribe(fn func(Event)) (id int) {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()
	c.subSeq++
	c.subscribers[c.subSeq] = fn
	return c.subSeq
}

func (c *Client) Unsubscribe(id int) {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()
	delete(c.subscribers, id)
}

// Call invokes a core method and returns the raw result.
func (c *Client) Call(method string, arguments any, timeout time.Duration) (json.RawMessage, error) {
	args := json.RawMessage("null")
	if arguments != nil {
		encoded, err := json.Marshal(arguments)
		if err != nil {
			return nil, err
		}
		args = encoded
	}
	c.mu.Lock()
	if c.conn == nil || c.closed {
		c.mu.Unlock()
		return nil, ErrClosed
	}
	c.nextID++
	id := strconv.Itoa(c.nextID)
	ch := make(chan methodResponse, 1)
	c.pending[id] = ch
	conn := c.conn
	c.mu.Unlock()

	call := methodCall{ID: id, Method: method, Arguments: args}
	data, err := json.Marshal(call)
	if err != nil {
		return nil, err
	}
	if err := writeFrame(conn, data); err != nil {
		c.resolve(id, methodResponse{})
		return nil, fmt.Errorf("core write: %w", err)
	}
	if timeout == 0 {
		timeout = callTimeout
	}
	select {
	case response := <-ch:
		if response.Error != nil {
			return nil, &MethodError{Code: response.Error.Code, Message: response.Error.Message}
		}
		if response.File != "" {
			return c.readSpilled(response.File)
		}
		return response.Result, nil
	case <-time.After(timeout):
		c.resolve(id, methodResponse{})
		return nil, errors.New("core call timeout: " + method)
	}
}

func (c *Client) readSpilled(path string) (json.RawMessage, error) {
	defer os.Remove(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func writeFrame(conn net.Conn, data []byte) error {
	if len(data) > maxFrameSize {
		return errors.New("frame too large")
	}
	buf := make([]byte, 4+len(data))
	binary.LittleEndian.PutUint32(buf[:4], uint32(len(data)))
	copy(buf[4:], data)
	conn.SetWriteDeadline(time.Now().Add(15 * time.Second))
	_, err := conn.Write(buf)
	conn.SetWriteDeadline(time.Time{})
	return err
}

// MethodError mirrors the core's MethodError on the Dart side.
type MethodError struct {
	Code    string
	Message string
}

func (e *MethodError) Error() string { return e.Message }

// Connected reports whether the core transport is up.
func (c *Client) Connected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil && !c.closed
}
