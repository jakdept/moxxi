package main

import (
	"errors"
)

// assign a unique id to each error
const (
	ErrCloseFile = 1 << iota
	ErrRemoveFile
)

// specify the error message for each error
const errMsg = map[int]string{
	ErrCloseFile:  "failed to close the file [%s] - %v",
	ErrRemoveFile: "failed to remove file [%s] - %v",
}

// declare the error type
type LocErr struct {
	ErrCode int
	Args    []struct{}
}

// the function `Error` to make my custom errors work
func (e *LocErr) Error() string {
	return fmt.Sprintf(errMsg[e.ErrCode], e.Args...)
}
