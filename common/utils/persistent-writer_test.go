package utils

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type WriterTestSuite struct {
	suite.Suite
}

func (w *WriterTestSuite) SetupSuite() {
	persistChannelPageSize = 10
	persistChannelWriteBuffer = 5
	getTestFilePath = func(page int64) string {
		return fmt.Sprintf("%s/writer_test.%v", testChannelDir, page)
	}
}

func (w *WriterTestSuite) TearDownSuite() {
	removeTestFiles()
}

func (w *WriterTestSuite) SetupTest() {
	removeTestFiles()
	setupTestData()
}

// TestWrite tests writer write data to channel.
func (w *WriterTestSuite) TestWrite() {
	writer := NewPersistChannelWriter(getTestFilePath)
	writer.page = 0
	writer.offset = 0
	writer.OpenPage()
	writer.Start()

	var cursor int64
	// Test tick flush.
	writer.Write([]byte(testData[cursor]))
	cursor++

	// Wait disk sync.
	time.Sleep(persistChannelBufferDuration * 100)
	require.Equal(w.T(), int64(0), writer.bufferLen)
	require.Equal(w.T(), cursor, writer.offset)
	assertPage(w.T(), 0, testData[:cursor])

	// Test buffer full flush.
	for i := int64(0); i < persistChannelWriteBuffer; i++ {
		writer.Write([]byte(testData[cursor]))
		cursor++
	}
	require.Equal(w.T(), int64(0), writer.bufferLen)
	require.Equal(w.T(), cursor, writer.offset)
	assertPage(w.T(), 0, testData[:cursor])

	// Test page full flush.
	remainSize := persistChannelPageSize - cursor
	for i := int64(0); i < remainSize; i++ {
		writer.Write([]byte(testData[cursor]))
		cursor++
	}
	require.Equal(w.T(), int64(0), writer.bufferLen)
	require.Equal(w.T(), cursor, writer.offset)
	assertPage(w.T(), 0, testData[:cursor])

	// Test 2 page write.
	for i := int64(0); i < persistChannelPageSize*2; i++ {
		writer.Write([]byte(testData[cursor]))
		cursor++
	}
	require.Equal(w.T(), int64(0), writer.bufferLen)
	require.Equal(w.T(), cursor, writer.offset)
	assertPage(
		w.T(),
		1,
		testData[persistChannelPageSize:persistChannelPageSize*2])
	assertPage(
		w.T(),
		2,
		testData[persistChannelPageSize*2:persistChannelPageSize*3])

	writer.Close()
}

// TestLargeWrite tests if writer can write big data.
func (w *WriterTestSuite) TestLargeWrite() {
	writer := NewPersistChannelWriter(getTestFilePath)
	writer.page = 0
	writer.offset = 0
	writer.OpenPage()
	writer.Start()

	data := strings.Repeat("G", 1024*1024*20)

	writer.Write([]byte(data))

	// Wait disk sync.
	time.Sleep(persistChannelBufferDuration * 100)
	assertPage(w.T(), 0, []string{data})

	writer.Close()
}

func TestWriter(t *testing.T) {
	suite.Run(t, new(WriterTestSuite))
}
