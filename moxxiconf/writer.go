package moxxiConf

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/dchest/uniuri"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func inArr(a []string, t string) bool {
	for _, s := range a {
		if t == s {
			return true
		}
	}
	return false
}

func validHost(s string) string {
	s = strings.Trim(s, ".")
	parts := strings.Split(s, DomainSep)
	if len(parts) < 2 {
		return ""
	}
	for i := 0; i < len(parts)-1; {
		switch {
		case len(parts[i]) < 1:
			parts = append(parts[:i], parts[i+1:]...)
		case isNotAlphaNum.MatchString(parts[i]):
			return ""
		default:
			i++
		}
	}
	return strings.Join(parts, DomainSep)
}

func confCheck(proxy siteParams, config HandlerConfig) (siteParams, Err) {
	var conf siteParams
	if conf.IntHost = validHost(proxy.IntHost); conf.IntHost == "" {
		return siteParams{}, &NewErr{Code: ErrBadHost, value: proxy.IntHost}
	}

	tempIP := net.ParseIP(proxy.IntIP)
	if tempIP == nil {
		return siteParams{}, &NewErr{Code: ErrBadIP, value: proxy.IntIP}
	}
	if len(config.ipList) > 0 && !ipListContains(tempIP, config.ipList) {
		return siteParams{}, &NewErr{Code: ErrBlockedIP, value: tempIP.String()}
	}

	conf.IntPort = 80
	if proxy.IntPort > 0 && proxy.IntPort < MaxAllowedPort {
		conf.IntPort = proxy.IntPort
	}

	conf.IntIP = tempIP.String()
	conf.Encrypted = proxy.Encrypted
	conf.StripHeaders = proxy.StripHeaders

	if config.redirectTracing {
		newIntHost, newIntPort, newEncrypted, err := redirectTrace(conf.IntHost, conf.IntPort, conf.Encrypted)
		if err == nil {
			conf.IntHost = newIntHost
			conf.IntPort = newIntPort
			conf.Encrypted = newEncrypted
		}
	}

	return conf, nil
}

func confWrite(config HandlerConfig) func(siteParams) (siteParams, Err) {

	return func(siteConfig siteParams) (siteParams, Err) {

		err := os.ErrExist
		var randPart, fileName string
		var f *os.File

		for os.IsExist(err) {
			randPart = uniuri.NewLenChars(config.subdomainLen, SubdomainChars)
			// pick again if you got something reserved
			if inArr(config.exclude, randPart) {
				continue
			}
			if inArr(config.exclude, randPart+DomainSep+config.baseURL) {
				continue
			}
			fileName = strings.Join([]string{
				strings.TrimRight(config.confPath, PathSep),
				PathSep,
				randPart,
				DomainSep,
				config.baseURL,
				DomainSep,
				strings.TrimLeft(config.confExt, DomainSep)}, "")
			f, err = os.Create(fileName)
		}

		siteConfig.ExtHost = strings.Join([]string{
			randPart,
			DomainSep,
			config.baseURL}, "")

		if err == os.ErrPermission {
			return siteParams{ExtHost: randPart}, &NewErr{Code: ErrFilePerm, value: fileName, deepErr: err}
		} else if err != nil {
			return siteParams{ExtHost: randPart}, &NewErr{Code: ErrFileUnexpect, value: fileName, deepErr: err}
		}

		tErr := config.confTempl.Execute(f, siteConfig)

		if err = f.Close(); err != nil {
			return siteParams{}, &NewErr{Code: ErrCloseFile, value: fileName, deepErr: err}
		}

		if tErr != nil {
			if err = os.Remove(fileName); err != nil {
				return siteParams{}, &NewErr{Code: ErrRemoveFile, value: fileName, deepErr: err}
			}
		}

		return siteConfig, nil
	}
}

func parseCheckbox(in string) bool {
	checkedValues := []string{
		"true",
		"checked",
		"on",
		"yes",
		"y",
		"1",
	}

	for _, each := range checkedValues {
		if each == in {
			return true
		}
	}
	return false
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

func ipListContains(address net.IP, list []*net.IPNet) bool {
	for _, each := range list {
		if each.Contains(address) {
			return true
		}
	}
	return false
}

func redirectTrace(initHost string, initPort int, initTLS bool) (string, int, bool, Err) {

	var initURL string
	switch {
	case initTLS && initPort == 443:
		initURL = (fmt.Sprintf("https://%s/", initHost))
	case initTLS && initPort != 443:
		initURL = (fmt.Sprintf("https://%s:%d/", initHost, initPort))
	case !initTLS && initPort == 80:
		initURL = (fmt.Sprintf("http://%s/", initHost))
	case !initTLS && initPort != 80:
		initURL = (fmt.Sprintf("http://%s:%d/", initHost, initPort))
	}

	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := c.Head(initURL)
	if err != nil {
		return "", 0, false, NewErr{
			Code:    ErrBadHostnameTrace,
			value:   initURL,
			deepErr: err,
		}
	}

	var respHost string
	var respPort int
	var respTLS bool

	if resp.Request == nil {
		return "", 0, false, NewErr{
			Code:    ErrBadHostnameTrace,
			value:   initURL,
			deepErr: fmt.Errorf("did not get a request back from %s", initURL),
		}
	}

	var hostname string
	if resp.Request.Host != "" {
		hostname = resp.Request.Host
	} else if resp.Request.URL != nil {
		hostname = resp.Request.URL.Host
	} else {
		return "", 0, false, NewErr{
			Code:    ErrBadHostnameTrace,
			value:   initURL,
			deepErr: fmt.Errorf("cound not find the URL"),
		}
	}

	if strings.Contains(hostname, ":") {
		parts := strings.Split(hostname, ":")
		respPort, err = strconv.Atoi(parts[len(parts)-1])
		if err != nil {
			respHost = resp.Request.Host
		} else {
			respHost = strings.Join(parts[:len(parts)-1], ":")
		}
	} else {
		respHost = hostname
		if resp.TLS == nil {
			respPort = 80
		} else {
			respPort = 443
		}
	}
	if resp.TLS != nil {
		respTLS = true
	}

	return respHost, respPort, respTLS, nil
}
