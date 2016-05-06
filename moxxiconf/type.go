package moxxiConf

import (
	"fmt"
	"strings"
	"time"
)

// PathSep is the path seperator used throughout this program
const PathSep = "/"

// DomainSep is the seperator used between subdomains
const DomainSep = "."

// DefaultBackendTLS is the default value to use for TLS
const DefaultBackendTLS = false

// ConnTimeout is the tiemout to use on the server
const ConnTimeout = 10 * time.Second

// MaxAllowedPort is the maximum allowed destination port
const MaxAllowedPort = 65535

var SubdomainChars = []byte("abcdeefghijklmnopqrstuvwxyz")

type siteParams struct {
	ExtHost      string
	IntHost      string
	IntIP        string
	IntPort      int
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
	default:
		return fmt.Sprintf(errMsg[e.Code], e.value, e.deepErr)
	}
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
	ErrCloseFile:    "failed to close the file [%s] - %v",
	ErrRemoveFile:   "failed to remove file [%s] - %v",
	ErrFilePerm:     "permission denied to create file [%s] - %v",
	ErrFileUnexpect: "unknown error with file [%s] - %v",
	ErrBadHost:      "bad hostname provided [%s]",
	ErrBadIP:        "bad IP provided [%s]",
	ErrNoRandom:     "was not given a new random domain - shutting down",
}

// HandlerLocFlag gives a built in way to specify multiple locations to put the same handler
type HandlerLocFlag []string

func (f HandlerLocFlag) String() string {
	switch {
	case len(f) < 1:
		return ""
	case len(f) < 2:
		return f[0]
	default:
		return strings.Join(f, " ")
	}
}

func (f *HandlerLocFlag) Set(value string) error {
	if strings.HasPrefix(value, PathSep) {
		*f = append(*f, value)
	}
	return nil
}

func (f HandlerLocFlag) GetOne(i int) string {
	return f[i]
}

func parseCheckbox(in string) bool {
	checkedValues := []string{
		"true",
		"checked",
		"on",
		"1",
	}

	for _, each := range checkedValues {
		if each == in {
			return true
		}
	}
	return false;
}