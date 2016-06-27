package moxxiConf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerLocFlag(t *testing.T) {
	var testData = []string{
		"/one",
		"/two",
		"three",
		"/four",
	}

	var expected = []string{
		"/one",
		"/two",
		"/four",
	}

	testWork := new(HandlerLocFlag)

	for _, each := range testData {
		err := testWork.Set(each)
		assert.NoError(t, err, "there should not have been a problem adding an item")
	}

	assert.Equal(t, "/one /two /four", testWork.String(), "the test input and current value of the test should be equal")

	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], testWork.GetOne(i), "one item from the test was incorrect")
	}

	junkTest := new(HandlerLocFlag)
	assert.Equal(t, "", junkTest.String(), "should be empty")
	junkTest.Set("/some/real/junk")
	assert.Equal(t, "/some/real/junk", junkTest.String(), "should be empty")

}

func TestIsNotAlphaNum(t *testing.T) {
	testData := []struct {
		in  string
		out bool
	}{
		{
			in:  "abcde",
			out: false,
		}, {
			in:  "12345",
			out: false,
		}, {
			in:  "abcd123",
			out: false,
		}, {
			in:  "a$c&1^3",
			out: true,
		},
	}

	for id, each := range testData {
		assert.Equal(t, each.out, isNotAlphaNum.MatchString(each.in),
			"test #%d - results did not match expected - input %s", id, each.in)
	}

}
