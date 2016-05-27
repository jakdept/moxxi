package moxxiConf

import (
	"github.com/dchest/uniuri"
	"net"
	"os"
	"strings"
	"text/template"
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

func confCheck(host, ip string, destTLS bool, port int, blockedHeaders []string) (siteParams, error) {
	var conf siteParams
	if conf.IntHost = validHost(host); conf.IntHost == "" {
		return siteParams{}, &Err{Code: ErrBadHost, value: host}
	}

	tempIP := net.ParseIP(ip)
	if tempIP == nil {
		return siteParams{}, &Err{Code: ErrBadIP, value: ip}
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

func confWrite(config HandlerConfig) func(siteParams) (siteParams, error) {

	return func(siteConfig siteParams) (siteParams, error) {

		err := os.ErrExist
		var randPart, fileName string
		var f *os.File

		for os.IsExist(err) {
			randPart = uniuri.NewLenChars(config.subdomainLen, SubdomainChars)
			// pick again if you got something reserved
			if inArr(config.excludes, randPart) {
				continue
			}
			if inArr(config.excludes, randPart+DomainSep+config.baseURL) {
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

		siteConfig.ExtHost = randPart

		if err == os.ErrPermission {
			return siteParams{ExtHost: randPart}, &Err{Code: ErrFilePerm, value: fileName, deepErr: err}
		} else if err != nil {
			return siteParams{ExtHost: randPart}, &Err{Code: ErrFileUnexpect, value: fileName, deepErr: err}
		}

		tErr := config.confTempl.Execute(f, siteConfig)

		if err = f.Close(); err != nil {
			return siteParams{}, &Err{Code: ErrCloseFile, value: fileName, deepErr: err}
		}

		if tErr != nil {
			if err = os.Remove(fileName); err != nil {
				return siteParams{}, &Err{Code: ErrRemoveFile, value: fileName, deepErr: err}
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
