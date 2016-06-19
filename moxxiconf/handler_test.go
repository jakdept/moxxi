package moxxiConf

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestFormHandler_POST(t *testing.T) {

	// test setup
	testConfig := HandlerConfig{
		baseURL:      "test.com",
		confPath:     os.TempDir(),
		confExt:      ".testout",
		exclude:      []string{"a", "b", "c"},
		subdomainLen: 8,
	}

	confTemplVal := "{{.IntHost}} {{.IntIP}} {{.IntPort}} {{.Encrypted}}"
	confTemplVal += " {{ range .StripHeaders }}{{.}} {{end}}"

	testConfig.confTempl = template.Must(template.New("testing").Parse(confTemplVal))

	resTemplVal := "{{range .}} {{ .ExtHost }} {{ end }}"
	testConfig.resTempl = template.Must(template.New("testing").Parse(resTemplVal))

	server := httptest.NewServer(FormHandler(testConfig,
		log.New(os.Stdout, "", log.LstdFlags)))
	defer server.Close()

	var testData = []struct {
		// reqMethod string
		reqParams map[string][]string
		resCode   int
		fileOut   string
	}{
		{
			reqParams: map[string][]string{
				"host":   []string{"proxied.com"},
				"ip":     []string{"10.10.10.10"},
				"port":   []string{"80"},
				"tls":    []string{"true"},
				"header": []string{"KeepAlive", "b", "c"},
			},
			resCode: 200,
			fileOut: `proxied.com 10.10.10.10 80 true KeepAlive b c `,
		}, {
			reqParams: map[string][]string{
				"ip":     []string{"10.10.10.10"},
				"port":   []string{"80"},
				"tls":    []string{"true"},
				"header": []string{"KeepAlive", "b", "c"},
			},
			resCode: http.StatusPreconditionFailed,
			fileOut: "no provided hostname\n",
		}, {
			reqParams: map[string][]string{
				"host":   []string{"proxied.com"},
				"port":   []string{"80"},
				"tls":    []string{"true"},
				"header": []string{"KeepAlive", "b", "c"},
			},
			resCode: http.StatusPreconditionFailed,
			fileOut: "no provided IP\n",
		}, {
			reqParams: map[string][]string{
				"host":   []string{"proxied.com"},
				"ip":     []string{"10.10.10.10"},
				"tls":    []string{"true"},
				"header": []string{"KeepAlive", "b", "c"},
			},
			resCode: 200,
			fileOut: `proxied.com 10.10.10.10 80 true KeepAlive b c `,
		}, {
			reqParams: map[string][]string{
				"host":   []string{".com"},
				"ip":     []string{"10.potato10.10.10"},
				"port":   []string{"80"},
				"tls":    []string{"true"},
				"header": []string{"KeepAlive", "b", "c"},
			},
			resCode: http.StatusPreconditionFailed,
			fileOut: "bad hostname provided [.com]\n",
		},
	}

	for id, test := range testData {
		params := url.Values(test.reqParams)
		resp, err := http.PostForm(server.URL, params)

		// req, err := http.NewRequest(test.reqMethod, server.URL,
		// 	strings.NewReader(params.Encode()))

		// resp, err := client.Do(req)

		assert.Equal(t, test.resCode, resp.StatusCode,
			"test %d - got the wrong response code", id)
		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err, "test %d - problem reading response - %v", id, err)
		if resp.StatusCode == 200 {
			proxyOut, err := ioutil.ReadFile(
				fmt.Sprintf("%s/%s%s",
					testConfig.confPath,
					bytes.TrimSpace(body),
					testConfig.confExt))

			assert.Nil(t, err, "test %d - problem reading file - %v", id, err)

			assert.Equal(t, test.fileOut, string(proxyOut),
				"test %d - wrong data written to the file", id)
		} else {
			assert.Equal(t, string(body), test.fileOut,
				"test %d - response and expected response did not match", id)
		}
		resp.Body.Close()
	}
}

func TestStaticHandler(t *testing.T) {
	// test setup
	expected := []byte(`this is the response I expect to recieve`)

	file, err := ioutil.TempFile(os.TempDir(), "moxxi_test_")
	assert.Nil(t, err, "could no open temp file for writing - %v", err)

	_, err = file.Write(expected)
	assert.Nil(t, err, "could no open temp file for writing - %v", err)

	server := httptest.NewServer(StaticHandler(HandlerConfig{resFile: file.Name()},
		log.New(os.Stdout, "", log.LstdFlags)))
	defer server.Close()

	for i := 0; i < 10; i++ {
		resp, err := http.Get(server.URL)
		assert.Nil(t, err, "got a bad response from the server - %v", err)

		actual, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err, "got an error reading the body of the response - %v", err)

		assert.Equal(t, expected, actual, "test #%d - got a different response than expected", i)
	}
}
