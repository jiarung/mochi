package utils

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ReaderTestSuite struct {
	suite.Suite
}

func (r *ReaderTestSuite) SetupSuite() {
	persistChannelPageSize = 10
	persistChannelRetainSize = 5
	getTestFilePath = func(page int64) string {
		return fmt.Sprintf("%s/reader_test.%v", testChannelDir, page)
	}
}

func (r *ReaderTestSuite) TearDownSuite() {
	removeTestFiles()
}

func (r *ReaderTestSuite) SetupTest() {
	setupTestData()

	var (
		f   *os.File
		err error
	)
	for idx, v := range testData {
		if int64(idx)%persistChannelPageSize == 0 {
			if f != nil {
				f.Close()
			}
			f, err = os.OpenFile(getTestFilePath(int64(idx)/persistChannelPageSize),
				os.O_SYNC|os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
			require.Nil(r.T(), err)
		}
		buffer := []byte(v)
		buffer = append(buffer, '\n')
		_, err = f.Write(buffer)
		require.Nil(r.T(), err)
	}
	f.Close()
}

// TestRead tests if reader read at the middle of channel.
func (r *ReaderTestSuite) TestRead() {
	var offset int64 = 13
	reader := NewPersistChannelReader(getTestFilePath, offset)
	reader.OpenPage()
	require.Equal(r.T(), offset, reader.offset)
	require.Equal(r.T(), offset/persistChannelPageSize, reader.page)

	for i := offset; i < offset+persistChannelPageSize; i++ {
		readData, readOffset := reader.Read()
		require.Equal(r.T(), i, readOffset)
		require.Equal(r.T(), testData[i], string(readData))
	}

	reader.Close()
}

// TestRead tests if reader read at the end of channel.
func (r *ReaderTestSuite) TestReadEnd() {
	offset := testDataSize
	reader := NewPersistChannelReader(getTestFilePath, offset)
	reader.OpenPage()
	require.Equal(r.T(), offset, reader.offset)
	require.Equal(r.T(), offset/persistChannelPageSize, reader.page)

	readData, readOffset := reader.Read()
	require.Equal(r.T(), offset, readOffset)
	require.Equal(r.T(), 0, len(readData))

	readData, readOffset = reader.Read()
	require.Equal(r.T(), offset, readOffset)
	require.Equal(r.T(), 0, len(readData))

	reader.Close()
}

// TestLargeRead tests if reader can read big data.
func (r *ReaderTestSuite) TestLargeRead() {
	f, err := os.OpenFile(getTestFilePath(0), os.O_SYNC|os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	require.Nil(r.T(), err)

	data := strings.Repeat("G", 1024*1024*20)
	_, err = f.Write([]byte(data))
	require.Nil(r.T(), err)
	_, err = f.Write([]byte{'\n'})
	require.Nil(r.T(), err)

	reader := NewPersistChannelReader(getTestFilePath, -1)
	reader.OpenPage()
	readData, readOffset := reader.Read()
	require.Equal(r.T(), int64(0), readOffset)
	require.Equal(r.T(), data, string(readData))
	reader.Close()
}

// TestRetain tests main reader should delete old pages.
func (r *ReaderTestSuite) TestRetain() {
	reader := NewPersistChannelReader(getTestFilePath, testDataSize)
	reader.OpenPage()

	retainPage := (testDataSize / persistChannelPageSize) - persistChannelRetainSize
	require.True(r.T(), retainPage >= 0)
	_, err := os.Stat(getTestFilePath(retainPage))
	require.NotNil(r.T(), err)
	require.True(r.T(), os.IsNotExist(err))

	reader.Close()
}

// TestShadow tests shadow reader should not delete old pages.
func (r *ReaderTestSuite) TestShadow() {
	reader := NewPersistChannelReader(getTestFilePath, testDataSize)
	reader.shadow = true
	reader.OpenPage()

	retainPage := (testDataSize / persistChannelPageSize) - persistChannelRetainSize
	require.True(r.T(), retainPage >= 0)
	_, err := os.Stat(getTestFilePath(retainPage))
	require.Nil(r.T(), err)

	reader.Close()
}

func TestReader(t *testing.T) {
	suite.Run(t, new(ReaderTestSuite))
}
