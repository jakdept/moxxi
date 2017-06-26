package moxxiConf

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestStaticHandler(t *testing.T) {
	// test setup
	expected := []byte(`this is the response I expect to recieve`)

	file, err := ioutil.TempFile(os.TempDir(), "moxxi_test_")
	assert.Nil(t, err, "could no open temp file for writing - %v", err)

	_, err = file.Write(expected)
	assert.Nil(t, err, "could no open temp file for writing - %v", err)

	server := httptest.NewServer(StaticHandler(HandlerConfig{resFile: file.Name()},
		log.New(ioutil.Discard, "", log.LstdFlags)))
	defer server.Close()

	for i := 0; i < 10; i++ {
		resp, err := http.Get(server.URL)
		assert.Nil(t, err, "got a bad response from the server - %v", err)

		actual, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err, "got an error reading the body of the response - %v", err)

		assert.Equal(t, expected, actual, "test #%d - got a different response than expected", i)
	}
}

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
		log.New(ioutil.Discard, "", log.LstdFlags)))
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
		assert.NoError(t, err, "test %d - got an error I should not have when running the request", id)
		if err != nil {
			continue
		}

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
		// resp.Body.Close()
	}
}

func TestJSONHandler_POST(t *testing.T) {

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

	resTemplVal := `{{ define "start" }}{{ end }}
	{{define "body" }}{{ .ExtHost }}
	{{ end }}
	{{ define "end" }}{{ end }}"`
	testConfig.resTempl = template.Must(template.New("testing").Parse(resTemplVal))

	server := httptest.NewServer(JSONHandler(testConfig,
		log.New(ioutil.Discard, "", log.LstdFlags)))
	defer server.Close()

	var testData = []struct {
		// reqMethod string
		reqBody string
		resCode int
		fileOut []string
	}{
		{
			reqBody: `{	"IntHost": "proxied.com", "IntIP": "10.10.10.10", "IntPort": 80,
				"Encrypted": true, "StripHeaders": [	"KeepAlive", "b", "c" ]}`,
			resCode: 200,
			fileOut: []string{
				`proxied.com 10.10.10.10 80 true KeepAlive b c `,
			},
		},
	}

	for id, test := range testData {
		resp, err := http.Post(server.URL, "application/json", strings.NewReader(test.reqBody))
		assert.NoError(t, err, "test %d - got an error I should not have when running the request", id)
		if err != nil {
			continue
		}

		assert.Equal(t, test.resCode, resp.StatusCode,
			"test %d - got the wrong response code", id)

		if resp.StatusCode == 200 {
			allFiles := test.fileOut

			s := bufio.NewScanner(resp.Body)
			for s.Scan() {
				fileName := strings.TrimSpace(s.Text())
				if fileName == "" {
					continue
				}
				contents, err := ioutil.ReadFile(
					fmt.Sprintf("%s/%s%s",
						testConfig.confPath,
						fileName,
						testConfig.confExt))

				assert.NoError(t, err, "test %d - problem reading file [%s] - %v", id, fileName, err)

				var found bool
				for i := 0; i < len(allFiles); i++ {
					if allFiles[i] == string(contents) {
						if i < len(allFiles)-1 {
							allFiles = append(allFiles[:i], allFiles[i+1:]...)
							found = true
							break
						} else {
							allFiles = allFiles[:i]
							found = true
							break
						}
					}
				}
				if !found {
					assert.Fail(t, "test %d - response not expected - opened file [%s]", id, fileName)
				}
			}

			if len(allFiles) > 0 {
				assert.Fail(t, "test %d - had results left over that were not found\n%v", id, allFiles)
			}
		}
		resp.Body.Close()
	}
}
