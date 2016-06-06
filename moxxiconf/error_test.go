package moxxiConf

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErr_Error(t *testing.T) {
	fakeError := errors.New("fake error")
	var testData = []struct {
		in  Err
		out string
	}{
		{
			NewErr{ErrCloseFile, "/tmp/testfile", fakeError},
			"failed to close the file [/tmp/testfile] - fake error",
		}, {
			NewErr{ErrRemoveFile, "/tmp/testfile", fakeError},
			"failed to remove file [/tmp/testfile] - fake error",
		}, {
			NewErr{ErrFilePerm, "/tmp/testfile", fakeError},
			"permission denied to create file [/tmp/testfile] - fake error",
		}, {
			NewErr{ErrFileUnexpect, "/tmp/testfile", fakeError},
			"unknown error with file [/tmp/testfile] - fake error",
		}, {
			NewErr{ErrBadHost, "/tmp/testfile", nil},
			"bad hostname provided [/tmp/testfile]",
		}, {
			NewErr{ErrBadIP, "/tmp/testfile", nil},
			"bad IP provided [/tmp/testfile]",
		}, {
			NewErr{ErrNoRandom, "", nil},
			"was not given a new random domain - shutting down",
		},
	}
	for _, test := range testData {
		testOut := test.in.Error()
		assert.Equal(t, test.out, testOut, "errors should match")
	}
}
