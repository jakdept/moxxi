package moxxiConf

import (
	"bufio"
	"net"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// PathSep is the path seperator used throughout this program
const PathSep = "/"

// DomainSep is the seperator used between subdomains
const DomainSep = "."

// DefaultBackendTLS is the default value to use for TLS
const DefaultBackendTLS = false

// ConnTimeout is the tiemout to use on the server
const ConnTimeout = 10 * time.Second

// MaxAllowedPort is the maximum allowed destination port
const MaxAllowedPort = 65535

var SubdomainChars = []byte("abcdeefghijklmnopqrstuvwxyz")

type siteParams struct {
	ExtHost      string
	IntHost      string
	IntIP        string
	IntPort      int
	Encrypted    bool
	StripHeaders []string
}

var isNotAlphaNum *regexp.Regexp

func init() {
	isNotAlphaNum = regexp.MustCompile("[^a-zA-Z0-9]")
}

type HandlerConfig struct {
	handlerType  string
	handlerRoute string
	baseURL      string
	confPath     string
	confExt      string
	exclude      []string
	confFile     string
	confTempl    *template.Template
	resFile      string
	resTempl     *template.Template
	ipFile       string
	ipList       []*net.IPNet
	subdomainLen int
}

func ipListContains(address net.IP, list []*net.IPNet) bool {
	for _, each := range list {
		if each.Contains(address) {
			return true
		}
	}
	return false
}

// everything below this line can likely go?
// #TODO#

// HandlerLocFlag gives a built in way to specify multiple locations to put the same handler
type HandlerLocFlag []string

func (f HandlerLocFlag) String() string {
	switch {
	case len(f) < 1:
		return ""
	case len(f) < 2:
		return f[0]
	default:
		return strings.Join(f, " ")
	}
}

func (f *HandlerLocFlag) Set(value string) error {
	if strings.HasPrefix(value, PathSep) {
		*f = append(*f, value)
	}
	return nil
}

func (f HandlerLocFlag) GetOne(i int) string {
	return f[i]
}

func parseIPList(ipFile string) ([]*net.IPNet, Err) {
	file, err := os.Open(ipFile)
	if err != nil {
		return []*net.IPNet{}, NewErr{
			Code:    ErrConfigBadIPFile,
			value:   ipFile,
			deepErr: err,
		}
	}

	var out []*net.IPNet

	s := bufio.NewScanner(file)

	for s.Scan() {
		t := strings.TrimSpace(s.Text())
		switch {
		case strings.HasPrefix(t, "//"):
		case strings.HasPrefix(t, "#"):
		case strings.HasPrefix(t, ";"):
		default:
			_, ipNet, err := net.ParseCIDR(s.Text())
			if err == nil {
				out = append(out, ipNet)
			}
		}
	}
	return out, nil
}
