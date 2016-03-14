package moxxiConf

import (
	"fmt"
	"strings"
)

// PathSep is the path seperator used throughout this program
const PathSep = "/"

// DomainSep is the seperator used between subdomains
const DomainSep = "."

// DefaultBackendTLS is the default value to use for TLS
const DefaultBackendTLS = false

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
	switch {
	case e.deepErr == nil && e.value == "":
		return errMsg[e.Code]
	case e.deepErr == nil && e.value != "":
		return fmt.Sprintf(errMsg[e.Code], e.value)
	case e.deepErr != nil && e.value == "":
		return fmt.Sprintf(errMsg[e.Code], e.deepErr)
	case e.deepErr != nil && e.value != "":
		return fmt.Sprintf(errMsg[e.Code], e.value, e.deepErr)
	}
	return nil
}

// assign a unique id to each error
const (
	ErrCloseFile = 1 << iota
	ErrRemoveFile
	ErrFilePerm
	ErrFileUnexpect
	ErrBadHost
	ErrBadIP
	ErrNoRandom
)

// specify the error message for each error
var errMsg = map[int]string{
	ErrCloseFile:    "failed to close the file [%s] - %s",
	ErrRemoveFile:   "failed to remove file [%s] - %s",
	ErrFilePerm:     "permission denied to create file [%s] - %s",
	ErrFileUnexpect: "unknown error with file [%s] - %s",
	ErrBadHost:      "bad hostname provided [%s]",
	ErrBadIP:        "bad IP provided [%s]",
	ErrNoRandom:     "was not given a new random domain - shutting down?",
}

// HandlerLocFlag gives a built in way to specify multiple locations to put the same handler
type HandlerLocFlag []string

func (f *HandlerLocFlag) String() string {
	return strings.Join(*f, " ")
}

func (f *HandlerLocFlag) Set(value string) error {
	for _, path := range strings.Split(value, ",") {
		if strings.HasPrefix(path, PathSep) {
			*f = append(*f, path)
		}
	}
	return nil
}
