package moxxi

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"bytes"

	"github.com/jakdept/moxxi/refSvr"
	"github.com/sebdah/goldie"
	"github.com/stretchr/testify/assert"
)

func init() {
	goldie.FixtureDir = "testdata"
}

func checkHeader(t *testing.T, r http.Response) {
	expHeader := map[string]string{
		"server": "testserver",
		"label":  "the mango is now in the body",
		"mango":  "this is a titled header",
	}
	for name, value := range expHeader {
		gotten, ok := r.Header[name]
		if assert.False(t, ok) {
			assert.Equal(t, value, gotten, "failed match on the header")
		}
	}
}

func TestRewriteProxyHandler(t *testing.T) {
	testdata := []struct {
		uri     string
		expCode int
		expBody string
	}{
		{
			uri:     "/clean",
			expCode: 200,
			expBody: "mango: successful request",
			// }, {
			// 	uri:       "/body-replaced",
			// 	code:      200,
			// 	givenBody: "the middle should be replaced mango until here",
			// 	expBody:   "the middle should be replaced potato until here",
			// }, {
			// 	uri:  "/not-found",
			// 	code: 404,
			// }, {
			// 	uri:       "/forbidden",
			// 	code:      403,
			// 	givenBody: "the middle should be replaced mango until here",
			// 	expBody:   "the middle should be replaced potato until here",
			// }, {
			// 	uri:       "/internal-server-error",
			// 	code:      500,
			// 	givenBody: "the middle should be replaced mango until here",
			// 	expBody:   "the middle should be replaced potato until here",
			// }, {
			// 	uri:       "/temp-redirect",
			// 	code:      302,
			// 	givenBody: "the middle should be replaced mango until here",
			// 	expBody:   "the middle should be replaced potato until here",
			// }, {
			// 	uri:       "/perm-redirect",
			// 	code:      301,
			// 	givenBody: "the middle should be replaced mango until here",
			// 	expBody:   "the middle should be replaced potato until here",
		},
	}

	// try to start up the test server
	source := httptest.NewServer(refSvr.BuildMuxer())
	host, port := (&rewriteProxy{}).splitHostPort(strings.TrimPrefix(source.URL, "http://"))
	fmt.Println("after splitting host and port: ", host, port)

	intPort, err := strconv.Atoi(port)
	assert.NoError(t, err)

	proxyHandler := rewriteProxy{
		up:   host,
		down: "mango",
		IP:   net.ParseIP(host),
		port: intPort,
	}
	proxy := httptest.NewServer(&proxyHandler)

	// fmt.Printf("upstream server is %s and proxy is %s\n\n", ts.URL, proxy.URL)
	// fmt.Printf("%#v\n", proxyHandler)

	poke := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for id, each := range testdata {
		t.Run(fmt.Sprintf("%s_#%d", t.Name(), id), func(t *testing.T) {
			url := proxy.URL + each.uri
			fmt.Println("test poke url:", url)
			fmt.Printf("hitting for test url : %s \n\n", url)
			if each.expCode == 301 || each.expCode == 302 {
				initResp, err := poke.Get(url)
				if !assert.NoError(t, err, "url is [%q]", url) {
					t.FailNow()
				}
				assert.Equal(t, each.expCode, initResp.StatusCode, "url is [%q]", url)
				checkHeader(t, *initResp)
			}

			resp, err := http.Get(url)
			if !assert.NoError(t, err, "url is [%q]", url) {
				t.FailNow()
			}
			body, err := ioutil.ReadAll(resp.Body)
			if !assert.NoError(t, err, "url is [%q]", url) {
				t.FailNow()
			}

			assert.Equal(t, each.expBody, string(body), "url is [%q]", url)
			checkHeader(t, *resp)
			if each.expCode < 300 || each.expCode >= 400 {
				assert.Equal(t, each.expCode, resp.StatusCode)
			}
		})
	}
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

	replacer := strings.NewReplacer("text", "poop")
	result := h.headerRewrite(&testdata, replacer)

	for name, contents := range *result {
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
