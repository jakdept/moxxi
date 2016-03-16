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
		if len(parts[i]) < 1 {
			parts = append(parts[:i], parts[i+1:]...)
		} else {
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

func confWrite(confPath, confExt, baseURL string, subdomainLen int, t template.Template,
	excludes []string) func(siteParams) (siteParams, error) {

	if subDomainLen < 1 {
		subdomainLen = 1
	}

	return func(config siteParams) (siteParams, error) {

		err := os.ErrExist
		var randPart, fileName string
		var f *os.File

		for randPart == "" || os.IsExist(err) {
			randPart = uniuri.NewLenChars(subdomainLen, SubdomainChars) + "." + baseURL
			// pick again
			if inArr(excludes, randPart) {
				continue
			}
			fileName = strings.TrimRight(confPath, PathSep) + PathSep
			fileName += randPart + DomainSep + strings.TrimLeft(confExt, DomainSep)
			f, err = os.Create(fileName)
		}

		config.ExtHost = randPart

		if err == os.ErrPermission {
			return siteParams{}, &Err{Code: ErrFilePerm, value: fileName, deepErr: err}
		} else if err != nil {
			return siteParams{}, &Err{Code: ErrFileUnexpect, value: fileName, deepErr: err}
		}

		tErr := t.Execute(f, config)

		if err = f.Close(); err != nil {
			return siteParams{}, &Err{Code: ErrCloseFile, value: fileName, deepErr: err}
		}

		if tErr != nil {
			if err = os.Remove(fileName); err != nil {
				return siteParams{}, &Err{Code: ErrRemoveFile, value: fileName, deepErr: err}
			}
		}

		return config, nil
	}
}
