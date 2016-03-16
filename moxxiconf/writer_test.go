package moxxiConf

import (
	"github.com/stretchr/testify/assert"
	"log"
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
	var testSize = []int{0, 1, 2, 3, 10}
	var testRepeat = 20
	var testDomain = "domain.com"
	var testExclude = []string{"lol.domain.com", "aaa.domain.com"}

	var done chan struct{}
	var count int

	expectedSize := testSize[0]
	if expectedSize < 2 {
		expectedSize = 2
	}
	expectedSize += len(testDomain)
	expectedSize++

	out := RandSeqFeeder(testDomain, testExclude, testSize[0], done)

	for count < testRepeat {
		log.Println("waiting for domain")
		log.Println(<-out)
	}

	done <- struct{}{}
	close(done)
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
