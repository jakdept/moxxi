package main

import (
	"github.com/JackKnifed/moxxi/moxxiconf"
	"log"
	"net/http"
)

func main() {
	config, err := moxxiConf.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	mux := moxxiConf.CreateMux(config)

	srv := http.Server{
		Addr:         config.Listen,
		Handler:      mux,
		ReadTimeout:  moxxiConf.ConnTimeout,
		WriteTimeout: moxxiConf.ConnTimeout,
	}

	log.Fatal(srv.ListenAndServe())
}
