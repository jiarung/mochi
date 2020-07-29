package utils

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ChannelTestSuite struct {
	suite.Suite
}

func (c *ChannelTestSuite) SetupSuite() {
	persistChannelPageSize = 10
	persistChannelWriteBuffer = 5
	getTestFilePath = func(page int64) string {
		return fmt.Sprintf("%s/%s.%v", testChannelDir, testChannelName, page)
	}
}

func (c *ChannelTestSuite) TearDownSuite() {
	removeTestFiles()
}

func (c *ChannelTestSuite) SetupTest() {
	removeTestFiles()
	setupTestData()

	var (
		f   *os.File
		err error
	)
	for idx, v := range testData {
		// only write half testData to file
		if int64(idx) == testDataSize/2 {
			break
		}
		if int64(idx)%persistChannelPageSize == 0 {
			if f != nil {
				f.Close()
			}
			f, err = os.OpenFile(getTestFilePath(int64(idx)/persistChannelPageSize),
				os.O_SYNC|os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
			require.Nil(c.T(), err)
		}
		buffer := []byte(v)
		buffer = append(buffer, '\n')
		_, err = f.Write(buffer)
		require.Nil(c.T(), err)
	}
	f.Close()
}

func (c *ChannelTestSuite) TestSetupOffset() {
	channel := NewPersistChannel(testChannelDir, testChannelName, 0)
	require.Equal(c.T(), int64(0), channel.reader.offset)
	require.Equal(c.T(), int64(0), channel.reader.page)
	require.Equal(c.T(), testDataSize/2, channel.writer.offset)
	require.Equal(
		c.T(),
		testDataSize/2/persistChannelPageSize,
		channel.writer.page)
	channel.Close()

	channel = NewPersistChannel(testChannelDir, testChannelName, -1)
	require.Equal(c.T(), testDataSize/2, channel.reader.offset)
	require.Equal(
		c.T(),
		testDataSize/2/persistChannelPageSize,
		channel.reader.page)
	require.Equal(c.T(), testDataSize/2, channel.writer.offset)
	require.Equal(
		c.T(),
		testDataSize/2/persistChannelPageSize,
		channel.writer.page)
	channel.Close()
}

func (c *ChannelTestSuite) TestIntegrityCheck() {
	os.Remove(getTestFilePath(0))
	require.Panics(c.T(), func() {
		NewPersistChannel(testChannelDir, testChannelName, 0)
	})
}

func (c *ChannelTestSuite) TestCommunicate() {
	channel := NewPersistChannel(testChannelDir, testChannelName, -1)
	require.Equal(c.T(), testDataSize/2, channel.reader.offset)
	require.Equal(
		c.T(),
		testDataSize/2/persistChannelPageSize,
		channel.reader.page)
	require.Equal(c.T(), testDataSize/2, channel.writer.offset)
	require.Equal(
		c.T(),
		testDataSize/2/persistChannelPageSize,
		channel.writer.page)

	channel.Write([]byte("123"))

	for {
		time.Sleep(persistChannelBufferDuration * 2)
		bytes, offset := channel.Read()
		if channel.reader.offset != offset {
			require.Equal(c.T(), "123", string(bytes))
			break
		}
	}

	channel.Close()
}

func (c *ChannelTestSuite) TestCreateShadowReader() {
	channel := NewPersistChannel(testChannelDir, testChannelName, 0)
	require.Equal(c.T(), int64(0), channel.reader.offset)
	require.Equal(c.T(), int64(0), channel.reader.page)
	require.Equal(c.T(), testDataSize/2, channel.writer.offset)
	require.Equal(
		c.T(),
		testDataSize/2/persistChannelPageSize,
		channel.writer.page)

	reader := channel.CreateShadowReader()
	require.True(c.T(), reader.shadow)
	require.Equal(c.T(), testDataSize/2, reader.offset)
	require.Equal(
		c.T(),
		testDataSize/2/persistChannelPageSize,
		reader.page)
	channel.Close()
	reader.Close()
}

func (c *ChannelTestSuite) TestSeekReader() {
	channel := NewPersistChannel(testChannelDir, testChannelName, -1)
	require.Equal(c.T(), testDataSize/2, channel.reader.offset)
	require.Equal(
		c.T(),
		testDataSize/2/persistChannelPageSize,
		channel.reader.page)
	require.Equal(c.T(), testDataSize/2, channel.writer.offset)
	require.Equal(
		c.T(),
		testDataSize/2/persistChannelPageSize,
		channel.writer.page)

	channel.Write([]byte("123"))
	time.Sleep(persistChannelBufferDuration * 2)
	bytes, offset := channel.Read()
	require.Equal(c.T(), "123", string(bytes))
	require.Equal(c.T(), testDataSize/2, offset)

	channel.SeekReader(offset)
	bytes, offset = channel.Read()
	require.Equal(c.T(), "123", string(bytes))
	require.Equal(c.T(), testDataSize/2, offset)

	channel.Close()
}

func TestChannel(t *testing.T) {
	suite.Run(t, new(ChannelTestSuite))
}
