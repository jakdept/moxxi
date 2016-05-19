package moxxiConf

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"text/template"
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
		}, {
			"sub.do_main.com",
			"",
		}, {
			"sub.do;main.com",
			"",
		},
	}
	for id, test := range testData {
		assert.Equal(t, test.out, validHost(test.in), "output should match - test # %d", id)
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
	var testData = []struct {
		in  siteParams
		out string
		err error
	}{
		{
			in: siteParams{
				IntHost:      "domain.com",
				IntIP:        "127.0.0.1",
				IntPort:      80,
				Encrypted:    true,
				StripHeaders: []string{"a", "b", "c"},
			},
			out: "domain.com 127.0.0.1 80 true a b c ",
			err: nil,
		}, {
			in: siteParams{
				IntHost:      "lol.com",
				IntIP:        "10.10.10.10",
				IntPort:      443,
				Encrypted:    true,
				StripHeaders: []string{"apple", "banana", "carrot"},
			},
			out: "lol.com 10.10.10.10 443 true apple banana carrot ",
			err: nil,
		},
	}

	templPath := os.TempDir()
	templExt := ".out"
	baseURL := "proxy.com"
	subdomainLen := 1
	excludes := []string{"a.domain.com", "b.domain.com", "c.domain.com"}
	templateString := `{{.IntHost}} {{.IntIP}} {{.IntPort}} {{.Encrypted}} {{ range .StripHeaders }}{{.}} {{end}}`
	templ := template.Must(template.New("testing").Parse(templateString))

	w := confWrite(templPath, templExt, baseURL, subdomainLen, *templ, excludes)

	var outConf siteParams
	var err, fileErr error
	var contents []byte

	for _, test := range testData {
		outConf, err = w(test.in)

		contents, fileErr = ioutil.ReadFile(templPath + PathSep + outConf.ExtHost + templExt)
		assert.Nil(t, fileErr, "problem reading file - %v", fileErr)

		assert.Equal(t, test.out, string(contents), "file contents did not match expected")
		assert.Equal(t, test.err, err, "errors did not match up")
	}
}

func TestConfWrite_badLocation(t *testing.T) {
	var test = struct {
		in  siteParams
		out string
		err error
	}{
		in: siteParams{
			IntHost:      "domain.com",
			IntIP:        "127.0.0.1",
			IntPort:      80,
			Encrypted:    true,
			StripHeaders: []string{"a", "b", "c"},
		},
	}

	templPath := "/bad/directory"
	templExt := ".out"
	baseURL := "proxy.com"
	subdomainLen := 1
	excludes := []string{"a.domain.com", "b.domain.com", "c.domain.com"}
	templateString := `{{.IntHost}} {{.IntIP}} {{.IntPort}} {{.Encrypted}} {{ range .StripHeaders }}{{.}} {{end}}`
	templ := template.Must(template.New("testing").Parse(templateString))

	w := confWrite(templPath, templExt, baseURL, subdomainLen, *templ, excludes)

	outConf, err := w(test.in)
	badFile := templPath + PathSep + outConf.ExtHost + templExt
	test.err = fmt.Errorf("unknown error with file [%s] - open %s: no such file or directory", badFile, badFile)

	assert.Equal(t, test.err.Error(), err.Error(), "errors did not match up")
}
