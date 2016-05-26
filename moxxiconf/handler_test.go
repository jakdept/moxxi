package moxxiConf

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"text/template"
)

func TestFormHandler_POST(t *testing.T) {
	t.SkipNow()
	var testData = []struct {
		reqMethod string
		reqURL    string
		reqParams map[string][]string
		resCode   int
		fileOut   string
	}{
		{
			reqMethod: "POST",
			reqURL:    "domain.com",
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

	templPath := os.TempDir()
	templExt := ".out"
	baseURL := "domain.com"
	subdomainLen := 1
	excludes := []string{"a.domain.com", "b.domain.com", "c.domain.com"}

	fileTemplString := `{{.IntHost}} {{.IntIP}} {{.IntPort}} {{.Encrypted}} {{ range .StripHeaders }}{{.}} {{end}}`
	fileTempl := template.Must(template.New("testing").Parse(fileTemplString))

	resTemplString := `{{ .ExtHost }}`
	resTempl := template.Must(template.New("testing").Parse(resTemplString))

	handler := FormHandler(baseURL, templPath, templExt, excludes, *fileTempl, *resTempl, subdomainLen)

	for _, test := range testData {
		w := httptest.NewRecorder()

		params := url.Values(test.reqParams)
		r, err := http.NewRequest(test.reqMethod, test.reqURL,
			strings.NewReader(params.Encode()))

		http.HandlerFunc(handler).ServeHTTP(w, r)

		log.Println(w.Body.String())

		file := templPath + PathSep + w.Body.String() + templExt
		contents, err := ioutil.ReadFile(file)

		assert.Nil(t, err, "problem reading file - %v", err)
		assert.Equal(t, test.resCode, w.Code, "got the wrong response code")
		assert.Equal(t, test.fileOut, string(contents), "wrong data written to the file")
	}
}
