package moxxiConf

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInArr(t *testing.T) {
	var testData = []struct {
		arr    []string
		target string
		out    bool
	}{
		{
			arr:    []string{"a", "b", "c"},
			target: "a",
			out:    true,
		}, {
			arr:    []string{"a", "b", "c"},
			target: "d",
			out:    false,
		}, {
			arr:    []string{"apple", "banana", "carrot"},
			target: "ana",
			out:    false,
		},
	}
	for _, testRun := range testData {
		result := inArr(testRun.arr, testRun.target)
		assert.Equal(t, testRun.out, result, "wrong response - expected %s", testRun.out)
	}
}

func TestRandSeqFeeder(t *testing.T) {
	var testData = []struct {
		in      int
		stopped bool
	}{
		{
			in:      0,
			stopped: false,
		}, {
			in:      1,
			stopped: false,
		}, {
			in:      2,
			stopped: true,
		}, {
			in:      3,
			stopped: true,
		}, {
			in:      10,
			stopped: true,
		}, {
			in:      20,
			stopped: true,
		},
	}
	var testRepeat = 20
	var testDomain = "domain.com"
	var exclude = []string{"lol.domain.com", "aaa.domain.com"}

	for _, test := range testData {
		var finisher chan struct{}
		out := RandSeqFeeder(testDomain, exclude, test.in, finisher)
		finalLength := len(testDomain) + test.in + 1

		finishedEarly := false
		var count int

		// go until we have the minimum number of interations
		// see if the thread closes itself, or if it feeds data back
		for count <= testRepeat {
			select {
			case s, more := <-out:
				if !more {
					finishedEarly = true
					close(finisher)
					break
				}
				assert.Equal(t, finalLength, len(s), "returned string should be the same length as the input")
				count++
			default:
				t.Fatal("response channel is delayed")
			}
		}

		// if the thread did not close early, and the count is correct, close the channel
		assert.Equal(t, testRepeat, count, "test did not run the correct amount of times")

		if testRepeat == count && !finishedEarly {
			close(finisher)
			select {
			case _, more := <-out:
				if more {
					t.Fatal("response channel should be closed, it is not")
				}
			default:
				t.Fatal("response channel is not closed and not responding")
			}
		}
		assert.Equal(t, test.stopped, finishedEarly, "the channel did not behave properly")
	}
}

func TestValidHost(t *testing.T) {
	var testData = []struct {
		in  string
		out string
	}{
		{
			"domain.com",
			"domain.com",
		}, {
			"sub.domain.com",
			"sub.domain.com",
		}, {
			".domain.com",
			"domain.com",
		}, {
			".sub.domain.com",
			"sub.domain.com",
		}, {
			"domain.com.",
			"domain.com",
		}, {
			".sub.domain.com.",
			"sub.domain.com",
		}, {
			"sub...domain.com",
			"sub.domain.com",
		}, {
			"...sub.domain.com...",
			"sub.domain.com",
		},
	}
	for _, test := range testData {
		assert.Equal(t, test.out, validHost(test.in), "output should match")
	}
}

func TestConfCheck(t *testing.T) {
	var testData = []struct {
		host, ip       string
		destTLS        bool
		port           int
		blockedHeaders []string
		exp            siteParams
		expErr         error
	}{
		{
			host:           "domain.com",
			ip:             "127.0.0.1",
			destTLS:        true,
			port:           80,
			blockedHeaders: []string{"a", "b", "c"},
			exp: siteParams{
				IntHost:      "domain.com",
				IntIP:        "127.0.0.1",
				IntPort:      80,
				Encrypted:    true,
				StripHeaders: []string{"a", "b", "c"},
			},
			expErr: nil,
		}, {
			host:           "com",
			ip:             "127.0.0.1",
			destTLS:        true,
			port:           80,
			blockedHeaders: []string{"a", "b", "c"},
			exp:            siteParams{},
			expErr:         &Err{Code: ErrBadHost, value: "com"},
		}, {
			host:           "domain.com",
			ip:             "127.1",
			destTLS:        true,
			port:           80,
			blockedHeaders: []string{"a", "b", "c"},
			exp:            siteParams{},
			expErr:         &Err{Code: ErrBadIP, value: "127.1"},
		},
	}

	for _, test := range testData {
		eachOut, eachErr := confCheck(test.host, test.ip, test.destTLS,
			test.port, test.blockedHeaders)
		assert.Equal(t, test.exp, eachOut, "expected return and actual did not match")
		assert.Equal(t, test.expErr, eachErr, "expected return and actual did not match")
	}
}

func TestConfWrite(t *testing.T) {

}
