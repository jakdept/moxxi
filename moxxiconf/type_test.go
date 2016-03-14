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
	for _, test := range testData {
		assert.Equal(t, test.out, test.in.Error(), "errors should match")
	}
}

func HandlerLocFlag_test(t *testing.T) {
	var testData = []string{
		"one",
		"two",
		"three",
		"four",
	}

	var testWork HandlerLocFlag

	for _, each := range testData {
		assert.NoError(t, testWork.Set(each), "there should not have been a problem adding an item")
	}

	assert.Equal(t, testData, []string(testWork), "the test input and current value of the test should be equal")

	for i := 0; i < len(testData); i++ {
		assert.Equal(t, testData[i], testWork.GetOne(i), "the test input and current value of the test should be equal")
	}

}
