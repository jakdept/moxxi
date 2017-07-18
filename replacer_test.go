package moxxi

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/sebdah/goldie"
	"github.com/stretchr/testify/assert"
)

func TestReplacerVanilla(t *testing.T) {
	t.Parallel()
	// should be no error
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
	done := r.Replace(file, buf)
	<-done
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
	done := r.Replace(file, buf)
	<-done
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
	done := r.Replace(file, buf)
	<-done
	goldie.Assert(t, t.Name(), buf.Bytes())
}

func TestReplacerEOF(t *testing.T) {
	t.Parallel()
	// should be no error
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

	junk := make([]byte, 2<<20)
	output := r.RewriteRequest(file)
	for i := 0; i < 10000; i++ {
		_, err = output.Read(junk)
		if err != nil {
			break
		}
	}

	assert.Equal(t, io.EOF, err)
}
