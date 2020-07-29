package fileio

import (
	"encoding/json"
	"os"
)

// DecodeStreamJSON read JSON object stream from given file.
// This could be used with `AppendJSON()` to save JSON objects and read one by
// one.
// File object is returned because the decoder would panic if file closes first.
func DecodeStreamJSON(fname string) (*json.Decoder, *os.File, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, nil, err
	}
	dec := json.NewDecoder(f)
	return dec, f, nil
}
