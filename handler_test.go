package moxxi

import (
	"os"
	"testing"

	"bytes"

	"github.com/sebdah/goldie"
	"github.com/stretchr/testify/assert"
)

func init() {
	goldie.FixtureDir = "testdata"
}

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

// func TestReplacerReversed(t *testing.T) {
// 	t.Parallel()
// 	defer func() {
// 		err, ok := recover().(error)
// 		if !assert.False(t, ok) {
// 			t.FailNow()
// 		}
// 		if !assert.NoError(t, err) {
// 			t.FailNow()
// 		}
// 	}()

// 	file, err := os.Open("testdata/lorem_ipsum.input")
// 	if !assert.NoError(t, err) {
// 		t.FailNow()
// 	}
// 	defer file.Close()
// 	r := Replacer{old: []byte("123"), new: []byte("cum")}
// 	r.Reverse()
// 	data := []byte{}
// 	buf := bytes.NewBuffer(data)
// 	r.replace(file, buf)
// 	goldie.Assert(t, t.Name(), data)
// }
