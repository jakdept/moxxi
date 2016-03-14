package moxxiConf

import (
	"github.com/dchest/uniuri"
	"net"
	"os"
	"strings"
	"text/template"
)

// persistently runs and feeds back random URLs.
// To be started concurrently.
func randSeqFeeder(baseURL, exclude string, length int, feeder chan<- string, done <-chan struct{}) {

	var chars = []byte("abcdeefghijklmnopqrstuvwxyz")
	defer close(feeder)
	//rand.Seed(time.New().UnixNano())

	var newURL string

	for {
		newURL = uniuri.NewLenChars(length, chars) + "." + baseURL
		if newURL == exclude {
			continue
		}
		select {
		case <-done:
			return
		case feeder <- newURL:
		}
	}
	}()

	return feeder, done
}

func validHost(s string) string {
	s = strings.Trim(s, ".")
	parts := len(strings.Split(s, "."))
	if parts < 2 {
		return ""
	}
	return s
}

func confCheck(host, ip string, destTLS bool, blockedHeaders []string) (siteParams, error) {
	var conf siteParams
	if conf.IntHost = validHost(host); conf.IntHost == "" {
		return siteParams{}, &Err{Code: ErrBadHost, value: ip}
	}

	tempIP := net.ParseIP(ip)
	if tempIP == nil {
		return siteParams{}, &Err{Code: ErrBadIP, value: ip}
	}

	conf.IntIP = tempIP.String()
	conf.Encrypted = destTLS
	conf.StripHeaders = blockedHeaders

	return conf, nil
}

func confWrite(confPath, confExt string, t template.Template,
	randHost <-chan string) func(siteParams) (string, error) {

	return func(config siteParams) (string, error) {

		err := os.ErrExist
		var randPart, fileName string
		var f *os.File

		for randPart == "" || os.IsExist(err) {
			select {
			case randPart = <-randHost:
			default:
				return "", &Err{Code: ErrNoRandom}
			}
			fileName = strings.TrimRight(confPath, PathSep) + PathSep
			fileName += randPart + DomainSep + mainDomain
			fileName += DomainSep + strings.TrimLeft(confExt, DomainSep)
			f, err = os.Create(fileName)
		}

		config.ExtHost = randPart + DomainSep + mainDomain

		if err == os.ErrPermission {
			return "", &Err{Code: ErrFilePerm, value: fileName, deepErr: err}
		} else if err != nil {
			return "", &Err{Code: ErrFileUnexpect, value: fileName, deepErr: err}
		}

		tErr := t.Execute(f, config)

		if err = f.Close(); err != nil {
			return "", &Err{Code: ErrCloseFile, value: fileName, deepErr: err}
		}

		if tErr != nil {
			if err = os.Remove(fileName); err != nil {
				return "", &Err{Code: ErrRemoveFile, value: fileName, deepErr: err}
			}
		}

		return config.ExtHost, nil
	}
}
