package main

import (
	"crypto/rand"
	"flag"
	"github.com/dchest/uniuri"
	"net/http"
	"net/url"
	"text/template"
)

type siteParams struct {
	ExtHost string
	IntHost string
	IntIP   net.IP
}

// persistently runs and feeds back random URLs.
// To be started concurrently.
func randSeqFeeder(baseURL string, length int, feeder chan<- string, done <-chan struct{}) {

	const chars = []bytes("abcdeefghijklmnopqrstuvwxyz")
	defer close(feeder)
	rand.Seed(time.New().UnixNano())
	newURL := urluri.NewLenChars(length, chars) + "." + baseURL

	for {
		select {
		case <-done:
			return
		case feeder <- newURL:
			newURL = urluri.NewLenChars(length, chars) + "." + baseURL
		}
	}
}

// TODO: standardize the errors in this function
func writeConf(config siteParams, confPath, confExt string, templ template.Template, randHost <-chan string) error {

	// set up the filename once
	config.ExtHost = <-randHost
	fileName := strings.TrimRight(confPath, os.PathSeperator)
	fileName += os.PathSeperator + ExtHost + confExt
	out, err := os.Create(fileName)

	// if you get an filename exists error, keep doing it until you don't
	for err != nil && os.IsExist(err) {
		config.ExtHost = <-randHost
		fileName = strings.TrimRight(confPath, os.PathSeperator)
		fileName += os.PathSeperator + ExtHost + confExt
		out, err = os.Create(fileName)
	}

	templErr = templ.Execute(out, config)

	if deepErr := f.Close(); deepErr != nil {
		return fmt.Errorf("failed close the file [%s] - %v,", err, fileName)
	}

	if templErr != nil {
		if deepErr := os.Remove(fileName); deepErr != nil {
			return fmt.Errorf("failed to create file - %v\nfailed to clean up [%s] - %v",
				err, fileName, deepErr)
		}
	}
}
