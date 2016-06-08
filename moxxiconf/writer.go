package moxxiConf

import (
	"github.com/dchest/uniuri"
	"net"
	"os"
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
	if !ipListContains(tempIP, ipList) {
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
