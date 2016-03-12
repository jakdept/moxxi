package main

import (
	"fmt"
)

// PathSep is the path seperator used throughout this program
const PathSep = "/"

type siteParams struct {
	ExtHost      string
	IntHost      string
	IntIP        string
	Encrypted    bool
	StripHeaders []string
}

// Err - the type used within my application for error handling
type Err struct {
	Code    int
	value   string
	deepErr error
}

// the function `Error` to make my custom errors work
func (e *Err) Error() string {
	if e.deepErr != nil {
		return fmt.Sprintf(errMsg[e.Code], e.value, e.deepErr)
	}
	return fmt.Sprintf(errMsg[e.Code], e.value)
}

// assign a unique id to each error
const (
	ErrCloseFile = 1 << iota
	ErrRemoveFile
	ErrFilePerm
	ErrFileUnexpect
	ErrBadHost
	ErrBadIP
)

// specify the error message for each error
var errMsg = map[int]string{
	ErrCloseFile:    "failed to close the file [%s] - %s",
	ErrRemoveFile:   "failed to remove file [%s] - %s",
	ErrFilePerm:     "permission denied to create file [%s] - %s",
	ErrFileUnexpect: "unknown error with file [%s] - %s",
	ErrBadHost:      "bad hostname provided [%s]",
	ErrBadIP:        "bad IP provided [%s]",
}
