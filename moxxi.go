package main

import (
	"github.com/JackKnifed/moxxi/moxxiconf"
	"log"
	"net/http"
)

func main() {
	listens, handlers, err := moxxiConf.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	mux := moxxiConf.CreateMux(handlers)

	var errChan chan error

	for _, singleListener := range listens {
		srv := http.Server{
			Addr:         singleListener,
			Handler:      mux,
			ReadTimeout:  moxxiConf.ConnTimeout,
			WriteTimeout: moxxiConf.ConnTimeout,
		}

		go func() {
			errChan <- srv.ListenAndServe()
		}()
	}

	log.Fatal(<-errChan)
}
