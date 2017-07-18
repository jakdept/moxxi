package moxxi

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/sebdah/goldie"
	"github.com/stretchr/testify/assert"
)

func TestReplacerVanilla(t *testing.T) {
	t.Parallel()
	defer func() {
		errInt := recover()
		if errInt != nil {
			err, ok := errInt.(error)
			if !assert.True(t, ok) {
				t.FailNow()
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}
		}
	}()

	file, err := os.Open("testdata/technologic.txt")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer file.Close()
	r := Replacer{
		old:      []byte(" it"),
		new:      []byte(" that"),
		memUsage: 50,
	}
	buf := &bytes.Buffer{}
	r.replace(file, buf)
	goldie.Assert(t, t.Name(), buf.Bytes())
}
func TestReplacerSmallBuffer(t *testing.T) {
	t.Parallel()
	defer func() {
		if err, ok := recover().(error); !ok {
			t.Error("recover did not return an error")
			t.FailNow()
		} else {
			if !assert.Equal(t, err,
				errors.New("not enough memory allocated for memory slice")) {
				t.FailNow()
			}
		}
	}()

	file, err := os.Open("testdata/technologic.txt")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer file.Close()
	r := Replacer{
		old:      []byte("really long replacement string"),
		new:      []byte("junk really"),
		memUsage: 1, // far too short for the buffer size
	}
	buf := &bytes.Buffer{}
	r.replace(file, buf)
}

func TestReplacerReversed(t *testing.T) {
	t.Parallel()
	defer func() {
		errInt := recover()
		if errInt != nil {
			err, ok := errInt.(error)
			if !assert.True(t, ok) {
				t.FailNow()
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}
		}
	}()

	file, err := os.Open("testdata/technologic.txt")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer file.Close()
	r := Replacer{
		old:      []byte(" that"),
		new:      []byte(" it"),
		memUsage: -1,
	}
	buf := &bytes.Buffer{}
	r.Reverse()
	r.replace(file, buf)
	goldie.Assert(t, t.Name(), buf.Bytes())
}
