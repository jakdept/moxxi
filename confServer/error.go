package main

import "fmt"

// LocErr - the type used within my application for error handling
type LocErr struct {
	ErrCode int
	Args    []interface{}
}

// the function `Error` to make my custom errors work
func (e *LocErr) Error() string {
	return fmt.Sprintf(errMsg[e.ErrCode], e.Args...)
}

// assign a unique id to each error
const (
	ErrCloseFile = 1 << iota
	ErrRemoveFile
)

// specify the error message for each error
var errMsg = map[int]string{
	ErrCloseFile:  "failed to close the file [%s] - %v",
	ErrRemoveFile: "failed to remove file [%s] - %v",
}
