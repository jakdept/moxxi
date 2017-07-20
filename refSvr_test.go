package moxxi

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func checkCleanHeader(t *testing.T, r http.Response) {
	expHeader := map[string]string{
		"Server": "testserver",
		"Refsvr": "this is a titled header",
		"Label":  "the refSvr is now in the body",
	}
	for name, value := range expHeader {
		gotten, ok := r.Header[name]
		if assert.True(t, ok,
			"the header is not there - %#v\n", r.Header) {
			assert.Equal(t, value, gotten[0], "failed match on the header")
		}
	}
}

func TestRefSvr(t *testing.T) {
	testdata := []struct {
		loc  string
		code int
		body string
	}{
		{
			loc:  "/",
			code: 200,
			body: "refSvr: successful request",
		}, {
			loc:  "/200",
			code: 200,
			body: "refSvr: successful request",
		}, {
			loc:  "/301",
			code: 301,
		}, {
			loc:  "/302",
			code: 302,
		}, {
			loc:  "/304",
			code: 304,
		}, {
			loc:  "/307",
			code: 307,
		}, {
			loc:  "/308",
			code: 308,
		}, {
			loc:  "/401",
			code: 401,
			body: "refSvr: authorization required\n",
		}, {
			loc:  "/403",
			code: 403,
			body: "refSvr: authorization required\n",
		}, {
			loc:  "/404",
			code: 404,
			body: "refSvr: not found\n",
		}, {
			loc:  "/500",
			code: 500,
			body: "refSvr: internal server error\n",
		}, {
			loc:  "/503",
			code: 503,
			body: "refSvr: gateway timeout\n",
		},
	}

	s := httptest.NewServer(AddHeaders(BuildRefSvrMuxer()))
	poke := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, test := range testdata {
		t.Run("TestRefSvr"+test.loc, func(t *testing.T) {
			t.Parallel()
			resp, err := poke.Get(s.URL + test.loc)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			assert.Equal(t, test.code, resp.StatusCode)
			checkCleanHeader(t, *resp)

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