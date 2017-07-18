package moxxi

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"net"

	"github.com/stretchr/testify/assert"
)

func TestStaticDialContext(t *testing.T) {
	testdata := []struct {
		url  string
		loc  string
		code int
		body string
	}{
		{
			url:  "http://google.com",
			loc:  "/",
			code: 200,
			body: "refSvr: successful request",
		}, {
			url:  "http://reddit.com",
			loc:  "/200",
			code: 200,
			body: "refSvr: successful request",
		}, {
			url:  "http://inbox.google.com",
			loc:  "/301",
			code: 301,
		}, {
			url:  "http://gmail.com",
			loc:  "/302",
			code: 302,
		}, {
			url:  "http://digitalocean.com",
			loc:  "/304",
			code: 304,
		}, {
			url:  "http://code.visualstudio.com",
			loc:  "/307",
			code: 307,
		}, {
			url:  "http://golang.org",
			loc:  "/308",
			code: 308,
		}, {
			url:  "http://godoc.org",
			loc:  "/401",
			code: 401,
			body: "refSvr: authorization required\n",
		}, {
			url:  "http://github.com",
			loc:  "/403",
			code: 403,
			body: "refSvr: authorization required\n",
		}, {
			url:  "http://gitlab.com",
			loc:  "/404",
			code: 404,
			body: "refSvr: not found\n",
		}, {
			url:  "http://amazon.com",
			loc:  "/500",
			code: 500,
			body: "refSvr: internal server error\n",
		}, {
			url:  "http://twitter.com",
			loc:  "/503",
			code: 503,
			body: "refSvr: gateway timeout\n",
		},
	}
	s := httptest.NewServer(AddHeaders(BuildRefSvrMuxer()))
	defer s.Close()

	strHost, strPort, err := net.SplitHostPort(
		strings.TrimPrefix(s.URL, "http://"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	ip := net.ParseIP(strHost)
	if !assert.NotEqual(t, nil, ip) {
		t.FailNow()
	}
	port, err := strconv.Atoi(strPort)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	dialer, err := StaticDialContext(ip, port)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	poke := http.Client{
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           dialer,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, test := range testdata {
		t.Run("TestRefSvr"+test.loc, func(t *testing.T) {
			resp, err := poke.Get(s.URL + test.loc)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			assert.Equal(t, test.code, resp.StatusCode)

			if test.body == "" {
				t.SkipNow()
			}
			actual, err := ioutil.ReadAll(resp.Body)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			assert.Equal(t, test.body, string(actual))
		})
	}

}
