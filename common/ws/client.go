// Copyright 2018 Cobinhood Inc. All rights reserved.

// Package ws provides wrapper around gorilla/websocket.
package ws

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// ErrNotConnected is returned when operating on some closed Client.
	ErrNotConnected = errors.New("not connected")
	// ErrAlreadyConnected is returned when starting already opened Client.
	ErrAlreadyConnected = errors.New("already connected")
)

// Client is a wraping of client side websocket with auto-reconnection.
type Client struct {
	// Interface parameters.
	URL     string
	Header  http.Header
	OnOpen  func(*Client)
	OnMsg   func(*Client, []byte)
	OnError func(*Client, error)
	OnClose func(*Client)

	// Internal paramters.
	ws     *websocket.Conn
	wsLock sync.Mutex

	// Status.
	open       bool
	statusLock sync.Mutex
	stopChan   chan struct{}
	doneChan   chan struct{}
}

// Replace websocket connection.
func (c *Client) setWS(ws *websocket.Conn) {
	// Protect ws value unchanged within lock.
	c.wsLock.Lock()
	defer c.wsLock.Unlock()
	c.ws = ws
}

// Handle a single connection until it dies.
func (c *Client) run() {
	// Ignore websocket handshake response.
	ws, _, err := websocket.DefaultDialer.Dial(c.URL, c.Header)
	if err != nil {
		if c.OnError != nil {
			c.OnError(c, err)
		}
		return
	}
	c.setWS(ws)
	// OnOpen, do some stuff like subscribe.
	if c.OnOpen != nil {
		c.OnOpen(c)
	}

	select {
	case <-c.stopChan:
		// Stopped, close websocket and goto OnClose logic.
		// This check need to be done after setWS so that Reconnect in Stop can
		// ensure killing the loop.
	default:
		// Start reading routine.
		for {
			mtype, msg, err := ws.ReadMessage()
			if err != nil {
				// OnErr
				if c.OnError != nil {
					c.OnError(c, err)
				}
				break
			}
			switch mtype {
			case websocket.PingMessage:
				// Handle ping message from server.
				c.send(websocket.PongMessage, nil)
			case websocket.TextMessage:
				// OnMsg, Handle msg.
				if c.OnMsg != nil {
					c.OnMsg(c, msg)
				}
			}
		}
	}
	ws.Close()
	c.setWS(nil)
	// OnClose, clean up.
	if c.OnClose != nil {
		c.OnClose(c)
	}
}

// Reconnect loop.
func (c *Client) runLoop() {
	for {
		c.run()
		select {
		case <-c.stopChan:
			close(c.doneChan)
			return
		default:
		}
		time.Sleep(time.Second)
	}
}

// closeWS closes current websocket connection if exists.
func (c *Client) closeWS() {
	c.wsLock.Lock()
	defer c.wsLock.Unlock()
	if c.ws != nil {
		// Close current socket. This is safe for concurrent calls.
		c.ws.Close()
	}
	// Else, current socket is already killed.
}

// Start websocket client.
func (c *Client) Start() error {
	c.statusLock.Lock()
	defer c.statusLock.Unlock()
	if c.open {
		return ErrAlreadyConnected
	}
	c.open = true
	c.stopChan = make(chan struct{})
	c.doneChan = make(chan struct{})
	go c.runLoop()
	return nil
}

// Reconnect forces the current websocket to disconnect and reconnect.
func (c *Client) Reconnect() {
	c.closeWS()
}

// Stop websocket client.
func (c *Client) Stop() error {
	c.statusLock.Lock()
	defer c.statusLock.Unlock()
	if !c.open {
		return ErrNotConnected
	}
	// Trigger stop. Avoid any new connection.
	close(c.stopChan)
	// Close websocket if already created.
	c.closeWS()
	// Wait until cleanup done.
	<-c.doneChan
	c.open = false
	return nil
}

func (c *Client) send(messageType int, data []byte) error {
	// Ensure only one writer.
	c.wsLock.Lock()
	defer c.wsLock.Unlock()
	if c.ws == nil {
		return ErrNotConnected
	}
	return c.ws.WriteMessage(messageType, data)
}

// Send message to server.
func (c *Client) Send(msg []byte) error {
	return c.send(websocket.TextMessage, msg)
}

// IsAlive returns liveness of websocket. False is returned if it is stopped or
// reconnecting.
func (c *Client) IsAlive() bool {
	c.wsLock.Lock()
	defer c.wsLock.Unlock()

	// This check should be atomic. In that case, we don't need lock though.
	return c.ws != nil
}
