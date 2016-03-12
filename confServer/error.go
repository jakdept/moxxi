package main

import "fmt"

// LocErr - the type used within my application for error handling
type LocErr struct {
	Code     int
	fileName string
	deepErr  error
}

// the function `Error` to make my custom errors work
func (e *LocErr) Error() string {
	return fmt.Sprintf(errMsg[e.Code], e.fileName, e.deepErr)
}

// assign a unique id to each error
const (
	ErrCloseFile = 1 << iota
	ErrRemoveFile
	ErrFilePerm
	ErrFileUnexpect
)

// specify the error message for each error
var errMsg = map[int]string{
	ErrCloseFile:    "failed to close the file [%s] - %s",
	ErrRemoveFile:   "failed to remove file [%s] - %s",
	ErrFilePerm:     "permission denied to create file [%s] - %s",
	ErrFileUnexpect: "unknown error with file [%s] - %s",
}
