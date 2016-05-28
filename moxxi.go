package main

import (
	"github.com/JackKnifed/moxxi/moxxiconf"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
)

func main() {

	err = moxxiconf.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	mux := moxxiconf.CreateMux(config)

	srv := http.Server{
		Addr:         *listen,
		Handler:      mux,
		ReadTimeout:  moxxiConf.ConnTimeout,
		WriteTimeout: moxxiConf.ConnTimeout,
	}

	log.Fatal(srv.ListenAndServe())
}
