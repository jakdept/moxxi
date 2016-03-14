package moxxiConf

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Err_Error_test(t *testing.T) {
	fakeError := errors.New("fake error")
	var testData = []struct {
		in  Err
		out string
	}{
		{
			Err{ErrCloseFile, "/tmp/testfile", fakeError},
			"failed to close the file [/tmp/testfile] - fake error",
		}, {
			Err{ErrRemoveFile, "/tmp/testfile", fakeError},
			"failed to remove file [/tmp/testfile] - fake error",
		}, {
			Err{ErrFilePerm, "/tmp/testfile", fakeError},
			"permission denied to create file [/tmp/testfile] - fake error",
		}, {
			Err{ErrFileUnexpect, "/tmp/testfile", fakeError},
			"unknown error with file [/tmp/testfile] - fake error",
		}, {
			Err{ErrBadHost, "/tmp/testfile", nil},
			"bad hostname provided [/tmp/testfile]",
		}, {
			Err{ErrFileUnexpect, "/tmp/testfile", nil},
			"bad IP provided [/tmp/testfile]",
		}, {
			Err{ErrFileUnexpect, "", nil},
			"was not given a new random domain - shutting down",
		},
	}
	for _, test := range testData{
		assert.Equal(t, test.out, test.in.Error(), "errors should match")
	}
}
