package utils

import (
	"context"
	"errors"
)

var (
	// ErrChannelFull defines error when channel is full.
	ErrChannelFull = errors.New("channel full")
)

// BufferedChannel defines buffered channel struct.
type BufferedChannel struct {
	*UnlimitedChannel
	err chan error
	cap uint64
}

// Err returns error channel.
func (c *BufferedChannel) Err() <-chan error {
	return c.err
}

// Cap returns capacity of buffered channel.
func (c *BufferedChannel) Cap() uint64 {
	return c.cap
}

func (c *BufferedChannel) sendErr(err error) {
	select {
	case c.err <- err:
	default:
	}
}

// NewBufferedChannel returns an buffered channel object.
func NewBufferedChannel(cap uint64) *BufferedChannel {
	ctx, cancel := context.WithCancel(context.Background())
	c := &BufferedChannel{
		UnlimitedChannel: &UnlimitedChannel{
			in:     make(chan interface{}),
			out:    make(chan interface{}),
			done:   make(chan struct{}),
			ctx:    ctx,
			cancel: cancel,
			deque:  NewDeque(),
		},
		err: make(chan error, 1),
		cap: cap,
	}
	go func() {
		defer func() {
			c.in = nil
			c.out = nil
			close(c.done)
		}()
		for {
			// Priority Done().
			select {
			case <-ctx.Done():
				return
			default:
			}

			out, ok := c.deque.Head()
			if !ok {
				select {
				case <-ctx.Done():
					return
				case in := <-c.in:
					if c.deque.Len() >= cap {
						c.sendErr(ErrChannelFull)
						continue
					}
					c.deque.PushBack(in)
				}
				continue
			}

			select {
			case <-ctx.Done():
				return
			case in := <-c.in:
				if c.deque.Len() >= cap {
					c.sendErr(ErrChannelFull)
					continue
				}
				c.deque.PushBack(in)
			case c.out <- out:
				c.deque.PopFront()
			}
		}
	}()
	return c
}
