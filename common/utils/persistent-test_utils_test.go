package utils

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	// for testing
	testData        []string
	testDataSize    int64 = 64
	testChannelName       = "channel_test"
	testChannelDir        = "/tmp"
	getTestFilePath func(int64) string
)

func assertPage(t *testing.T, page int64, expect []string) {
	path := getTestFilePath(page)
	bytes, err := ioutil.ReadFile(path)
	require.Nil(t, err)
	data := strings.Split(string(bytes), "\n")
	// remove tailing '\n'
	data = data[:len(data)-1]
	require.Equal(t, len(expect), len(data))
	for idx, s := range data {
		require.Equal(t, expect[idx], s)
	}
}

func removeTestFiles() {
	for i := int64(0); i <= testDataSize; i++ {
		if i%persistChannelPageSize == 0 {
			os.Remove(getTestFilePath(i / persistChannelPageSize))
		}
	}
}

func setupTestData() {
	testData = nil
	for _, v := range rand.Perm(int(testDataSize)) {
		s := fmt.Sprintf("%v", v)
		testData = append(testData, s)
	}
}
