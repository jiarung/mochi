package ws

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
}

type wsHandler struct {
	lock   sync.Mutex
	server *websocket.Conn

	read chan []byte
}

func newWSHandler() *wsHandler {
	return &wsHandler{
		read: make(chan []byte),
	}
}

func (h *wsHandler) connect(w http.ResponseWriter, r *http.Request) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.server != nil {
		// Only allow one concurrent connection in tests.
		return errors.New("multiple ws connected")
	}
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	h.server = c
	return nil
}

func (h *wsHandler) disconnect() {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.server.Close()
	h.server = nil
}

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.connect(w, r)
	if err != nil {
		return
	}
	defer h.disconnect()

	for {
		mt, message, err := h.server.ReadMessage()
		if err != nil {
			return
		}
		if mt == websocket.TextMessage {
			h.read <- message
		} // Ignore control messages.
	}
}

func (s *ClientTestSuite) TestCallbacks() {
	handler := newWSHandler()

	// Setup echo server for testing.
	server := httptest.NewServer(handler)
	defer server.Close()

	connectURL, err := url.Parse(server.URL)
	s.Require().Nil(err)
	// Change connect scheme.
	connectURL.Scheme = "ws"

	// Test the order of callbacks.
	cbEvents := []string{}
	openWait := make(chan struct{})

	c := Client{
		URL: fmt.Sprintf(connectURL.String()),
		OnOpen: func(c *Client) {
			openWait <- struct{}{}
			cbEvents = append(cbEvents, "open")
		},
		OnMsg: func(c *Client, msg []byte) {
			cbEvents = append(cbEvents, "msg")
			cbEvents = append(cbEvents, string(msg))
		},
		OnError: func(c *Client, err error) {
			cbEvents = append(cbEvents, "error")
			// Error content could be websocket/http/tcp closed error.
			// Hard to check unexported error, ignore for now.
		},
		OnClose: func(c *Client) {
			cbEvents = append(cbEvents, "close")
		},
	}

	writeMsg := "123"
	readMsg := "234"
	expectedEvents := []string{
		"open",         // Start.
		"msg", readMsg, // Echo from server.
		"error", "close", "open", // Reconnect1. (trigger by CloseMessage)
		"error", "close", "open", // Reconnect2. (call Reconnect)
		"error", "close", // Stop.
	}

	c.Start()
	<-openWait

	// Send msg to server.
	err = c.Send([]byte(writeMsg))
	s.Require().Nil(err)
	// Check msg is received.
	msg := <-handler.read
	s.Require().Equal([]byte(writeMsg), msg)

	// Send msg to read from server.
	err = handler.server.WriteMessage(websocket.TextMessage, []byte(readMsg))
	s.Require().Nil(err)
	time.Sleep(time.Second) // Wait message to reach client.

	// Trigger disconnect by sending close message.
	c.ws.WriteControl(websocket.CloseMessage, nil, time.Time{})
	<-openWait // Wait ws reconnect.

	// Force reconnect.
	c.Reconnect()
	<-openWait // Wait ws reconnect.

	c.Stop()

	s.Require().Equal(expectedEvents, cbEvents)
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
