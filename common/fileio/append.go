package fileio

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Shared / public constant.
const (
	AppendMode os.FileMode = 0666

	AppendFlag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
)

// Append obj to givne file. If file doest not exsit, automatically create one.
// And JSON serialization is supported. If flag is false, append with
// fmt.Sprintf("%v\n")
func Append(fname string, obj interface{}, isJSON bool) (err error) {
	path := filepath.Dir(fname)
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return
	}

	f, err := os.OpenFile(fname, AppendFlag, AppendMode)
	if err != nil {
		return
	}
	defer f.Close()

	if isJSON {
		enc := json.NewEncoder(f)
		err = enc.Encode(obj)
	} else {
		_, err = f.WriteString(fmt.Sprintf("%v\n", obj))
	}
	return
}

// AppendJSON is wrapper of Append with isJSON flag equals true.
func AppendJSON(fname string, obj interface{}) error {
	return Append(fname, obj, true)
}

// AppendLog is wrapper of Append with isJSON flag equals false.
func AppendLog(fname string, obj interface{}) error {
	return Append(fname, obj, false)
}
