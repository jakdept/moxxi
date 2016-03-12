package main

import (
	// "crypto/rand"
	// "flag"
	"github.com/dchest/uniuri"
	"os"
	"strings"
	// "net/http"
	// "net/url"
	"text/template"
)

type siteParams struct {
	ExtHost string
	IntHost string
	IntIP   string
}

// persistently runs and feeds back random URLs.
// To be started concurrently.
func randSeqFeeder(baseURL string, length int, feeder chan<- string, done <-chan struct{}) {

	var chars = []byte("abcdeefghijklmnopqrstuvwxyz")
	defer close(feeder)
	//rand.Seed(time.New().UnixNano())
	newURL := uniuri.NewLenChars(length, chars) + "." + baseURL

	for {
		select {
		case <-done:
			return
		case feeder <- newURL:
			newURL = uniuri.NewLenChars(length, chars) + "." + baseURL
		}
	}
}

func writeConf(config siteParams, confPath, confExt string, templ template.Template, randHost <-chan string) (string, error) {

	// set up the filename once
	config.ExtHost = <-randHost
	fileName := strings.TrimRight(confPath, PathSep)
	fileName += PathSep + config.ExtHost + confExt
	out, err := os.Create(fileName)

	// if you get an filename exists error, keep doing it until you don't
	for err == nil && os.IsExist(err) {
		config.ExtHost = <-randHost
		fileName = strings.TrimRight(confPath, PathSep)
		fileName += PathSep + config.ExtHost + confExt
		out, err = os.Create(fileName)
	}

	if err == os.ErrPermission {
		return "", &LocErr{Code: ErrFilePerm, fileName: fileName, deepErr: err}
	}

	if err != nil {
		return "", &LocErr{Code: ErrFileUnexpect, fileName: fileName, deepErr: err}
	}

	templErr := templ.Execute(out, config)

	if err = out.Close(); err != nil {
		return "", &LocErr{Code: ErrCloseFile, fileName: fileName, deepErr: err}
	}

	if templErr != nil {
		if err = os.Remove(fileName); err != nil {
			return "", &LocErr{Code: ErrRemoveFile, fileName: fileName, deepErr: err}
		}
	}

	return config.ExtHost, nil
}
