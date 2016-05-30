package moxxiConf

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/http"
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

type Err interface {
	error
	LogError(*http.Request) string
}

// Err - the type used within my application for error handling
type NewErr struct {
	Code    int
	value   string
	deepErr error
}

func UpgradeError(e error) Err {
	return NewErr{Code: ErrUpgradedError, deepErr: e}
}

// the function `Error` to make my custom errors work
func (e NewErr) Error() string {
	switch {
	case e.Code == ErrUpgradedError && e.value == "":
		return e.deepErr.Error()
	case e.deepErr == nil && e.value == "":
		return errMsg[e.Code]
	case e.deepErr == nil && e.value != "":
		return fmt.Sprintf(errMsg[e.Code], e.value)
	default:
		return fmt.Sprintf(errMsg[e.Code], e.value, e.deepErr)
	}
}

// the function `LogError` to print error log lines
func (e NewErr) LogError(r *http.Request) string {
	ts := time.Now()
	switch {
	case e.Code == ErrUpgradedError && e.value == "":
		return fmt.Sprintf("%s %s",
			ts.Format("02-Jan-2006:15:04:05-0700"),
			errMsg[e.Code])
	case e.deepErr == nil && e.value == "":
		return fmt.Sprintf("%s %s",
			ts.Format("02-Jan-2006:15:04:05-0700"),
			errMsg[e.Code])
	case e.deepErr == nil && e.value != "":
		return fmt.Sprintf("%s %s %s "+errMsg[e.Code],
			ts.Format("02-Jan-2006:15:04:05-0700"),
			r.RemoteAddr,
			r.RequestURI,
			e.value)
	default:
		return fmt.Sprintf("%s %s %s "+errMsg[e.Code],
			ts.Format("02-Jan-2006:15:04:05-0700"),
			r.RemoteAddr,
			r.RequestURI,
			e.value,
			e.deepErr)
	}
}

// assign a unique id to each error
const (
	ErrUpgradedError = 1 << iota
	ErrCloseFile
	ErrRemoveFile
	ErrFilePerm
	ErrFileUnexpect
	ErrBadHost
	ErrBadIP
	ErrNoRandom
	ErrNoHostname
	ErrNoIP
	ErrConfigBadHost
)

// specify the error message for each error
var errMsg = map[int]string{
	ErrUpgradedError: "not actually an error message",
	ErrCloseFile:     "failed to close the file [%s] - %v",
	ErrRemoveFile:    "failed to remove file [%s] - %v",
	ErrFilePerm:      "permission denied to create file [%s] - %v",
	ErrFileUnexpect:  "unknown error with file [%s] - %v",
	ErrBadHost:       "bad hostname provided [%s]",
	ErrBadIP:         "bad IP provided [%s]",
	ErrNoRandom:      "was not given a new random domain - shutting down",
	ErrNoHostname:    "no provided hostname",
	ErrNoIP:          "no provided IP",
	ErrConfigBadHost: "Bad hostname for handler [%s]",
}

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
	excludes     []string
	confFile     string
	confTempl    *template.Template
	resFile      string
	resTempl     *template.Template
	subdomainLen int
}

type MoxxiConf struct {
	Handlers []HandlerConfig
	Listen   string
}

func loadConfig() (MoxxiConf, error) {
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("Fatal error config file: %s \n", err)
		return MoxxiConf{}, err
	}

	var config MoxxiConf
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode config into struct, %v", err)
	}

	return config, nil
}

func LoadConfig() (MoxxiConf, error) {
	prepConfigDefaults()

	config, err := loadConfig()
	if err != nil {
		return MoxxiConf{}, err
	}

	if err = verifyConfig(&config); err != nil {
		return MoxxiConf{}, err
	}

	return config, nil
}

func prepConfigDefaults() {
	// establish the config paths
	viper.SetConfigName("config.json")
	viper.SetConfigName("config.yaml")
	viper.SetConfigName("config.toml")
	viper.AddConfigPath("/etc/moxxi/")
	viper.AddConfigPath("$HOME/.moxxi")
	viper.AddConfigPath(".")

	// set default values for the config
	viper.SetDefault("listen", ":8080")
}

func verifyConfig(c *MoxxiConf) Err {
	var err error
	for i := 1; i < len(c.Handlers); i++ {
		if c.Handlers[i].confFile != "" {
			c.Handlers[i].confTempl, err = template.ParseFiles(c.Handlers[i].confFile)
			if err != nil {
				return UpgradeError(err)
			}
		}
		if c.Handlers[i].resFile != "" && c.Handlers[i].handlerType != "static" {
			c.Handlers[i].resTempl, err = template.ParseFiles(c.Handlers[i].resFile)
			if err != nil {
				return UpgradeError(err)
			}
		}
		if validHost(c.Handlers[i].baseURL) != c.Handlers[i].baseURL {
			return &NewErr{Code: ErrConfigBadHost, value: c.Handlers[i].baseURL}
		}
		if c.Handlers[i].subdomainLen < 4 {
			c.Handlers[i].subdomainLen = 4
		}
	}
	return nil
}
