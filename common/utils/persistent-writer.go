package utils

import (
	"os"
	"sync"
	"time"
)

// PersistChannelWriter provides sequential write mechanism of persist channel.
type PersistChannelWriter struct {
	PersistChannelFileOperator
	// buffer
	buffer    []byte
	bufferLen int64
	// control channels
	// ctrl for main routine send command to write routine
	ctrl chan writeCtrlType
	// ackCtrl for write routine ack to main routine
	ackCtrl chan bool
	// ticker
	ticker *time.Ticker
	// mutex for buffer
	bufferMutex sync.Mutex
	// mutex for write
	mutex sync.Mutex
}

// NewPersistChannelWriter create a new persist channel writer.
func NewPersistChannelWriter(getPath func(int64) string) *PersistChannelWriter {
	return &PersistChannelWriter{
		PersistChannelFileOperator: PersistChannelFileOperator{
			getPath: getPath,
		},
		ctrl:    make(chan writeCtrlType),
		ackCtrl: make(chan bool),
	}
}

// Start starts the writer flush ticker.
func (w *PersistChannelWriter) Start() {
	w.ticker = time.NewTicker(persistChannelBufferDuration)
	go func() {
	mainLoop:
		for {
			select {
			case <-w.ticker.C:
				w.flush()
			case cmd := <-w.ctrl:
				switch cmd {
				case writeFlush:
					w.flush()
					w.ackCtrl <- true
				case writeClose:
					w.flush()
					w.ackCtrl <- true
					break mainLoop
				}
			}
		}
	}()
}

// OpenPage opens the next write page.
func (w *PersistChannelWriter) OpenPage() {
	if w.file != nil {
		w.file.Close()
	}
	f, err := os.OpenFile(w.getPath(w.page),
		os.O_SYNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	w.file = f
}

// Close closes the writer.
func (w *PersistChannelWriter) Close() {
	if w.ticker != nil {
		w.ctrl <- writeClose
		<-w.ackCtrl
	}
	if w.file != nil {
		w.file.Close()
		w.file = nil
	}
}

// Write writes data to writer's buffer,
// it will be flushed to page at next tick.
func (w *PersistChannelWriter) Write(data []byte) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.bufferMutex.Lock()
	w.bufferLen++
	data = append(data, '\n')
	w.buffer = append(w.buffer, data...)
	needFlush := (w.offset+w.bufferLen)%persistChannelPageSize == 0 ||
		w.bufferLen >= persistChannelWriteBuffer
	w.bufferMutex.Unlock()
	if needFlush {
		w.Flush()
	}
}

// Flush triggers flush action.
func (w *PersistChannelWriter) Flush() {
	w.ctrl <- writeFlush
	<-w.ackCtrl
}

func (w *PersistChannelWriter) flush() {
	w.bufferMutex.Lock()
	defer w.bufferMutex.Unlock()

	if w.bufferLen == 0 {
		return
	}

	buffer := w.buffer
	w.offset += w.bufferLen
	w.buffer = nil
	w.bufferLen = 0

	_, err := w.file.Write(buffer)
	if err != nil {
		panic(err)
	}

	err = w.file.Sync()
	if err != nil {
		panic(err)
	}

	if w.offset%persistChannelPageSize == 0 {
		w.page++
		w.OpenPage()
	}

	// reset ticker
	w.ticker.Stop()
	w.ticker = time.NewTicker(persistChannelBufferDuration)
}
