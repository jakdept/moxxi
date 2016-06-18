package moxxiConf

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

	testConfig.confTempl = template.Must(template.New("testing").Parse(
		`{{.IntHost}} {{.IntIP}} {{.IntPort}} {{.Encrypted}} {{ range .StripHeaders }}{{.}} {{end}}`))

	testConfig.resTempl = template.Must(template.New("testing").Parse(
		`{{range .}} {{ .ExtHost }} {{ end }}`))

	server := httptest.NewServer(FormHandler(testConfig))
	defer server.Close()

	var testData = []struct {
		// reqMethod string
		reqParams map[string][]string
		resCode   int
		fileOut   string
	}{
		{
			// reqMethod: "POST",
			reqParams: map[string][]string{
				"host":   []string{"proxied.com"},
				"ip":     []string{"10.10.10.10"},
				"port":   []string{"80"},
				"tls":    []string{"true"},
				"header": []string{"KeepAlive", "b", "c"},
			},
			resCode: 200,
			fileOut: `proxied.com 10.10.10.10 80 true KeepAlive b c `,
		},
	}

	for _, test := range testData {
		params := url.Values(test.reqParams)
		resp, err := http.PostForm(server.URL, params)

		// req, err := http.NewRequest(test.reqMethod, server.URL,
		// 	strings.NewReader(params.Encode()))

		// resp, err := client.Do(req)

		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err, "problem reading response - %v", err)

		proxyOut, err := ioutil.ReadFile(
			fmt.Sprintf("%s/%s%s",
				testConfig.confPath,
				bytes.TrimSpace(body),
				testConfig.confExt))

		assert.Nil(t, err, "problem reading file - %v", err)

		assert.Equal(t, test.resCode, resp.StatusCode, "got the wrong response code")
		assert.Equal(t, test.fileOut, string(proxyOut), "wrong data written to the file")
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

	server := httptest.NewServer(StaticHandler(HandlerConfig{resFile: file.Name()}))
	defer server.Close()

	for i := 0; i < 10; i++ {
		resp, err := http.Get(server.URL)
		assert.Nil(t, err, "got a bad response from the server - %v", err)

		actual, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err, "got an error reading the body of the response - %v", err)

		assert.Equal(t, expected, actual, "test #%d - got a different response than expected", i)
	}
}
