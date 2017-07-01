package moxxi

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"bytes"

	"github.com/sebdah/goldie"
	"github.com/stretchr/testify/assert"
)

func init() {
	goldie.FixtureDir = "testdata"
}

func TestHeaderRewrite(t *testing.T) {
	t.Parallel()
	testdata := http.Header{
		"Date":              []string{"Sat, 01 Jul 2017 21:17:52 GMT"},
		"Expires":           []string{"-1"},
		"Cache-Control":     []string{"private, max-age=0"},
		"text-Control":      []string{"public, max-age=100"},
		"Content-Type":      []string{"text/html; charset=ISO-8859-1"},
		"P3P":               []string{"CP=\"This is not a text policy!"},
		"Server":            []string{"gws"},
		"X-XSS-Protection":  []string{"1; mode=block"},
		"X-Frame-Options":   []string{"SAMEORIGIN"},
		"Alt-Svc":           []string{"quic=\":443\"; ma=2592000; v=\"39,38,37,36,35\""},
		"Transfer-Encoding": []string{"chunked"},
		"text-encoding":     []string{"chunked"},
		"Accept-Ranges":     []string{"text"},
		"Vary":              []string{"Accept-Encoding"},
	}

	expected := http.Header{
		"Date":              []string{"Sat, 01 Jul 2017 21:17:52 GMT"},
		"Expires":           []string{"-1"},
		"Cache-Control":     []string{"private, max-age=0"},
		"Poop-Control":      []string{"public, max-age=100"},
		"Content-Type":      []string{"poop/html; charset=ISO-8859-1"},
		"P3p":               []string{"CP=\"This is not a poop policy!"},
		"Server":            []string{"gws"},
		"X-Xss-Protection":  []string{"1; mode=block"},
		"X-Frame-Options":   []string{"SAMEORIGIN"},
		"Alt-Svc":           []string{"quic=\":443\"; ma=2592000; v=\"39,38,37,36,35\""},
		"Transfer-Encoding": []string{"chunked"},
		"Poop-Encoding":     []string{"chunked"},
		"Accept-Ranges":     []string{"poop"},
		"Vary":              []string{"Accept-Encoding"},
	}

	h := rewriteProxy{}
	result := http.Header{}

	replacer := strings.NewReplacer("text", "poop")
	h.headerRewrite(&testdata, &result, replacer)

	for name, contents := range result {
		assert.Equal(t, expected[name], contents, "key was %s", name)
	}
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
