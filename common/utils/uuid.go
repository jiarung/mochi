package utils

import (
	"bytes"
	"log"
	"sort"

	"github.com/satori/go.uuid"
)

// UUIDArray implements `Interface` in sort package.
type UUIDArray []uuid.UUID

func (arr UUIDArray) Len() int {
	return len(arr)

}

func (arr UUIDArray) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(arr[i].Bytes(), arr[j].Bytes()) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

func (arr UUIDArray) Swap(i, j int) {
	arr[j], arr[i] = arr[i], arr[j]
}

// SortUuids sorts uuids.
func SortUuids(src []uuid.UUID) UUIDArray {
	arr := UUIDArray(src)
	sort.Sort(arr)
	return src
}
