package moxxiConf

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"text/template"
)

func TestFormHandler_POST(t *testing.T) {
	var testData = []struct {
		reqMethod string
		reqParams map[string][]string
		resCode   int
		fileOut   string
	}{
		{
			reqMethod: "POST",
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
		`{{ .ExtHost }}`))

	server := httptest.NewServer(FormHandler(testConfig))
	defer server.Close()

	client := &http.Client{}

	for _, test := range testData {
		params := url.Values(test.reqParams)
		req, err := http.NewRequest(test.reqMethod, server.URL,
			strings.NewReader(params.Encode()))

		resp, err := client.Do(req)

		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err, "problem reading response - %v", err)

		proxyOut, err := ioutil.ReadFile(
			fmt.Sprintf("%s/%s.%s",
				testConfig.confPath,
				bytes.TrimSpace(body),
				testConfig.confExt))

		assert.Nil(t, err, "problem reading file - %v", err)

		assert.Equal(t, test.resCode, resp.StatusCode, "got the wrong response code")
		assert.Equal(t, test.fileOut, string(proxyOut), "wrong data written to the file")
		resp.Body.Close()
	}
}
