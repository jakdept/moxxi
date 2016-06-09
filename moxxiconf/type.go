package moxxiConf

import (
	"net"
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
