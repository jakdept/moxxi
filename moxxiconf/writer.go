package moxxiConf

import (
	"bufio"
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

func confCheck(host, ip string, destTLS bool, port int,
	blockedHeaders []string, ipList []*net.IPNet) (siteParams, Err) {
	var conf siteParams
	if conf.IntHost = validHost(host); conf.IntHost == "" {
		return siteParams{}, &NewErr{Code: ErrBadHost, value: host}
	}

	tempIP := net.ParseIP(ip)
	if tempIP == nil {
		return siteParams{}, &NewErr{Code: ErrBadIP, value: ip}
	}
	if len(ipList) > 0 && !ipListContains(tempIP, ipList) {
		return siteParams{}, &NewErr{Code: ErrBlockedIP, value: tempIP.String()}
	}

	conf.IntPort = 80
	if port > 0 && port < MaxAllowedPort {
		conf.IntPort = port
	}

	conf.IntIP = tempIP.String()
	conf.Encrypted = destTLS
	conf.StripHeaders = blockedHeaders

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

func redirectTrace(initURL string, initPort int) (string, int, Err) {
	c := &http.Client{}
	resp, err := c.Get(initURL + strconv.Itoa(initPort))
	if err == nil {
		return "", 0, NewErr{
			Code:    ErrBadHostnameTrace,
			value:   initURL,
			deepErr: err,
		}
	}

	var respHost string
	var respPort int

	if strings.Contains(resp.Request.Host, ":") {
		parts := strings.Split(resp.Request.Host, ":")
		respPort, err = strconv.Atoi(parts[len(parts)-1])
		if err != nil {
			respHost = resp.Request.Host
		} else {
			respHost = strings.Join(parts[:len(parts)-1], ":")
		}
	} else {
		respHost = resp.Request.Host
		if resp.TLS == nil {
			respPort = 80
		} else {
			respPort = 443
		}
	}

	return respHost, respPort, nil
}
