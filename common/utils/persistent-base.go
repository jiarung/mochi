package utils

import (
	"os"
	"time"
)

type writeCtrlType int

var (
	persistChannelPageSize       int64 = 10000
	persistChannelRetainSize     int64 = 5
	persistChannelWriteBuffer    int64 = 1000
	persistChannelBufferDuration       = time.Millisecond

	writeClose writeCtrlType
	writeFlush writeCtrlType = 1
)

// PersistChannelFileOperator is base of file operator in channel.
type PersistChannelFileOperator struct {
	getPath func(int64) string
	offset  int64
	page    int64
	file    *os.File
}

// GetOffset returns current offset.
func (o *PersistChannelFileOperator) GetOffset() int64 {
	return o.offset
}
