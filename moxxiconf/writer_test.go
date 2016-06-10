package moxxiConf

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"text/template"
	// "strings"
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
			expErr:         &NewErr{Code: ErrBadHost, value: "com"},
		}, {
			host:           "domain.com",
			ip:             "127.1",
			destTLS:        true,
			port:           80,
			blockedHeaders: []string{"a", "b", "c"},
			exp:            siteParams{},
			expErr:         &NewErr{Code: ErrBadIP, value: "127.1"},
		},
	}

	for _, test := range testData {
		vhostConf := siteParams{IntHost: test.host, IntPort: test.port, Encrypted: test.destTLS, IntIP: test.ip, StripHeaders: test.blockedHeaders}
		eachOut, eachErr := confCheck(vhostConf, HandlerConfig{})
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

	templateString := `{{.IntHost}} {{.IntIP}} {{.IntPort}} {{.Encrypted}} {{ range .StripHeaders }}{{.}} {{end}}`

	testConfig := HandlerConfig{
		baseURL:      "proxy.com",
		confPath:     os.TempDir(),
		confExt:      ".out",
		exclude:      []string{"a.domain.com", "b.domain.com", "c.domain.com"},
		confTempl:    template.Must(template.New("testing").Parse(templateString)),
		subdomainLen: 1,
	}

	w := confWrite(testConfig)

	var outConf siteParams
	var err, fileErr error
	var contents []byte

	for _, test := range testData {
		outConf, err = w(test.in)

		contents, fileErr = ioutil.ReadFile(testConfig.confPath + PathSep + outConf.ExtHost + testConfig.confExt)
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

	templateString := `{{.IntHost}} {{.IntIP}} {{.IntPort}} {{.Encrypted}} {{ range .StripHeaders }}{{.}} {{end}}`

	testConfig := HandlerConfig{
		baseURL:      "proxy.com",
		confPath:     "/bad/directory",
		confExt:      ".out",
		exclude:      []string{"a.domain.com", "b.domain.com", "c.domain.com"},
		confTempl:    template.Must(template.New("testing").Parse(templateString)),
		subdomainLen: 1,
	}

	w := confWrite(testConfig)

	outConf, err := w(test.in)
	badFile := testConfig.confPath
	badFile += PathSep
	badFile += outConf.ExtHost
	badFile += DomainSep
	badFile += testConfig.baseURL
	badFile += testConfig.confExt
	test.err = fmt.Errorf("unknown error with file [%s] - open %s: no such file or directory", badFile, badFile)

	assert.Equal(t, test.err.Error(), err.Error(), "errors did not match up")
}

func TestParseCheckbox(t *testing.T) {
	var testData = []struct {
		in  string
		out bool
	}{
		{"true", true},
		{"ture", false},
		{"false", false},
		{"checked", true},
		{"unchecked", false},
		{"on", true},
		{"off", false},
		{"yes", true},
		{"no", false},
		{"y", true},
		{"n", false},
		{"1", true},
		{"0", false},
		{"2", false},
	}

	for id, test := range testData {
		assert.Equal(t, test.out, parseCheckbox(test.in),
			"output not epected - test %d", id)
	}
}

func TestIPListContains(t *testing.T) {
	var ipRanges = []string{
		"127.0.0.1/8",
		"10.0.0.0/8",
		"192.168.0.0/16",
	}

	var ipList []*net.IPNet

	for _, cidr := range ipRanges {
		_, oneRange, err := net.ParseCIDR(cidr)
		assert.Nil(t, err, "failed to parse IP range %s", cidr)
		ipList = append(ipList, oneRange)
	}

	var testData = []struct {
		ip       string
		expected bool
	}{
		{"127.0.0.1", true},
		{"127.0.0.200", true},
		{"127.0.0.2", true},
		{"127.0.10.100", true},
		{"128.0.0.1", false},
		{"10.10.10.1", true},
		{"10.10.10.10", true},
		{"11.10.10.10", false},
		{"192.168.0.1", true},
		{"192.168.0.2", true},
		{"192.168.255.1", true},
		{"192.169.255.1", false},
	}

	for id, each := range testData {
		ip := net.ParseIP(each.ip)
		assert.NotNil(t, ip, "failed to parse IP %s", each.ip)
		assert.Equal(t, each.expected, ipListContains(ip, ipList),
			"test %d did not match up", id)
	}
}

func TestParseIPList(t *testing.T) {
	testIn := `
127.0.0.0/8
10.0.0.0/8
192.168.0.0/16
#8.8.8.8/32
;8.8.4.4
//4.4.4.4
`
	file, err := ioutil.TempFile(os.TempDir(), "moxxiConfTest")
	assert.Nil(t, err, "failed to open file - %v", err)
	fileName := file.Name()
	defer os.Remove(fileName)

	_, err = file.Write([]byte(testIn))
	assert.Nil(t, err, "failed to write file - %v", err)

	err = file.Close()
	assert.Nil(t, err, "failed to close file - %v", err)

	actualIPList, _ := parseIPList(fileName)

	var ipRanges = []string{
		"127.0.0.1/8",
		"10.0.0.0/8",
		"192.168.0.0/16",
	}

	var expectedIPList []*net.IPNet

	for _, cidr := range ipRanges {
		_, oneRange, err := net.ParseCIDR(cidr)
		assert.Nil(t, err, "failed to parse IP range %s", cidr)
		expectedIPList = append(expectedIPList, oneRange)
	}

	assert.Equal(t, expectedIPList, actualIPList,
		"got the wrong ipList")
}

func TestParseIPList_BadFile(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	assert.Nil(t, err, "failed to open file - %v", err)
	fileName := file.Name()
	file.Close()
	os.Remove(file.Name())

	_, locErr := parseIPList(fileName)
	assert.Equal(t, ErrConfigBadIPFile, locErr.GetCode(), "got the wrong error type back")
}

func TestRedirectTrace(t *testing.T) {
	var testData = []struct {
		hostIn  string
		portIn  int
		tlsIn bool
		hostOut string
		portOut int
		tlsOut bool
	}{
		{
			hostIn:  "google.com",
			portIn:  80,
			tlsIn: false,
			hostOut: "www.google.com",
			portOut: 80,
			tlsOut: false,
		}, {
			hostIn:  "github.com",
			portIn:  80,
			tlsIn: false,
			hostOut: "github.com",
			portOut: 443,
			tlsOut: true,
		}, {
			hostIn:  "facebook.com",
			portIn:  80,
			tlsIn: false,
			hostOut: "www.facebook.com",
			portOut: 443,
			tlsOut: true,
		},
	}

	for id, test := range testData {
		hostRes, portRes, tlsRes, err := redirectTrace(test.hostIn, test.portIn, test.tlsIn)
		assert.Nil(t, err,
			"test %d - got an error back that I should not have\n%v", id, err)
		assert.Equal(t, test.hostOut, hostRes,
			"test %d - got the wrong/unexpected host back", id)
		assert.Equal(t, test.portOut, portRes,
			"test %d - got the wrong/unexpected port back", id)
		assert.Equal(t, test.tlsOut, tlsRes,
			"test %d - got the wrong/unexpected encryption back", id)
	}
}
