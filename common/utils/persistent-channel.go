package utils

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

// PersistChannel provides an infinite channel with disk.
type PersistChannel struct {
	dir    string
	name   string
	writer *PersistChannelWriter
	reader *PersistChannelReader
	closed bool
}

// NewPersistChannel create a new persist channel instance.
func NewPersistChannel(dir, name string, roffset int64) *PersistChannel {
	channel := &PersistChannel{
		dir:  dir,
		name: name,
	}
	channel.reader = NewPersistChannelReader(channel.getPath, roffset)
	channel.writer = NewPersistChannelWriter(channel.getPath)

	channel.setupOffset()
	channel.writer.Start()
	channel.writer.OpenPage()
	channel.reader.OpenPage()
	return channel
}

func (c *PersistChannel) getPath(page int64) string {
	return fmt.Sprintf("%s/%s.%v", c.dir, c.name, page)
}

func (c *PersistChannel) integrityCheck() {
	if c.writer.page == 0 {
		return
	}
	for i := c.reader.page; i <= c.writer.page; i++ {
		if _, err := os.Stat(c.getPath(i)); os.IsNotExist(err) {
			c.Close()
			panic(err)
		}
	}
}

func (c *PersistChannel) setupOffset() {
	files, err := ioutil.ReadDir(c.dir)
	if err != nil {
		panic(err)
	}

	nameLen := len(c.name)
	for _, file := range files {
		name := file.Name()
		// find filname "{channel_name}.{page}"
		if len(name) > nameLen && name[:nameLen] == c.name {
			page, err := strconv.ParseInt(name[nameLen+1:], 10, 64)
			if err == nil && page > c.writer.page {
				c.writer.page = page
			}
		}
	}

	f, err := os.OpenFile(c.getPath(c.writer.page), os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(f)
	if reader == nil {
		panic("nil buffer reader")
	}

	c.writer.offset = c.writer.page * persistChannelPageSize
	for {
		_, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			panic(err)
		}
		if err == io.EOF {
			break
		}
		c.writer.offset++
	}

	f.Close()
	if c.reader.offset > c.writer.offset || c.reader.offset < 0 {
		c.reader.offset = c.writer.offset
		c.reader.page = c.reader.offset / persistChannelPageSize
	}
	c.integrityCheck()
}

// CreateShadowReader create a shadow reader with this channel.
// Shadow reader wouldn't retain pages.
func (c *PersistChannel) CreateShadowReader() *PersistChannelReader {
	reader := NewPersistChannelReader(c.getPath, c.writer.offset)
	reader.shadow = true
	reader.OpenPage()
	return reader
}

// IsClosed returns is the channel is closed.
func (c *PersistChannel) IsClosed() bool {
	return c.closed
}

// Close close the persist channel.
func (c *PersistChannel) Close() {
	c.closed = true
	c.writer.Close()
	c.reader.Close()
}

// SeekReader seeks reader to offset.
func (c *PersistChannel) SeekReader(offset int64) {
	c.reader.Close()
	if offset > c.writer.offset || offset < 0 {
		offset = c.writer.offset
	}
	c.reader = NewPersistChannelReader(c.getPath, offset)
	c.integrityCheck()
	c.reader.OpenPage()
}

// Flush flushs data in writer.
func (c *PersistChannel) Flush() {
	c.writer.Flush()
}

// Read reads data from reader.
func (c *PersistChannel) Read() ([]byte, int64) {
	if c.closed {
		return nil, c.reader.offset
	}
	return c.reader.Read()
}

// Write writes data to writer.
func (c *PersistChannel) Write(data []byte) {
	if c.closed {
		return
	}
	c.writer.Write(data)
}
