package utils

import (
	"bytes"
	"io"
	"os"
	"sync"
)

var readerBufferSize = 8192

// PersistChannelReader provides sequential read mechanism of persist channel.
type PersistChannelReader struct {
	PersistChannelFileOperator
	buffer []byte
	shadow bool
	lock   sync.Mutex
}

// NewPersistChannelReader create a new persist channel reader.
func NewPersistChannelReader(
	getPath func(int64) string,
	offset int64) *PersistChannelReader {
	return &PersistChannelReader{
		PersistChannelFileOperator: PersistChannelFileOperator{
			getPath: getPath,
			offset:  offset,
			page:    offset / persistChannelPageSize,
		},
	}
}

// OpenPage open the next read page.
func (r *PersistChannelReader) OpenPage() {
	if r.file != nil {
		r.file.Close()
	}
	f, err := os.OpenFile(r.getPath(r.page), os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	r.file = f
	r.buffer = make([]byte, 0)
	if !r.shadow {
		if r.page >= persistChannelRetainSize {
			os.Remove(r.getPath(r.page - persistChannelRetainSize))
		}
	}
	offset := r.offset
	// move cursor to correct position.
	r.offset = r.page * persistChannelPageSize
	for r.offset < offset {
		r.Read()
	}
}

// Close closes the reader.
func (r *PersistChannelReader) Close() {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.file != nil {
		r.file.Close()
		r.file = nil
		r.buffer = nil
	}
}

// Readline read a line.
func (r *PersistChannelReader) Readline() []byte {
	var result []byte

	if idx := bytes.IndexByte(r.buffer, '\n'); idx != -1 {
		result = r.buffer[:idx]
		r.buffer = append([]byte{}, r.buffer[idx+1:]...)
		return result
	}
	tmpBuffer := make([]byte, readerBufferSize)
	for {
		if r.file == nil {
			break
		}
		b, err := r.file.Read(tmpBuffer)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if err != nil || b == 0 {
			break
		}
		r.buffer = append(r.buffer, tmpBuffer[0:b]...)
		if idx := bytes.IndexByte(tmpBuffer[0:b], '\n'); idx != -1 {
			break
		}
	}
	if idx := bytes.IndexByte(r.buffer, '\n'); idx != -1 {
		result = r.buffer[:idx]
		r.buffer = append([]byte{}, r.buffer[idx+1:]...)
		return result
	}
	return nil
}

// Read reads data from page.
func (r *PersistChannelReader) Read() ([]byte, int64) {
	r.lock.Lock()
	defer r.lock.Unlock()

	bytes := r.Readline()
	if len(bytes) == 0 {
		return bytes, r.offset
	}
	r.offset++
	if r.offset%persistChannelPageSize == 0 {
		r.page++
		r.OpenPage()
	}
	return bytes, r.offset - 1
}
