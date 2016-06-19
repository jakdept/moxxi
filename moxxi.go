package main

import (
	"log"
	"net/http"
	"os"

	"github.com/JackKnifed/moxxi/moxxiconf"
	gorillaHandlers "github.com/gorilla/handlers"
)

func main() {
	listens, handlers, err := moxxiConf.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stderr, "", log.LstdFlags|log.LUTC|log.Lshortfile)
	mux := moxxiConf.CreateMux(handlers, logger)

	var errChan chan error

	for _, singleListener := range listens {
		srv := http.Server{
			Addr:         singleListener,
			Handler:      gorillaHandlers.LoggingHandler(os.Stdout, mux),
			ReadTimeout:  moxxiConf.ConnTimeout,
			WriteTimeout: moxxiConf.ConnTimeout,
		}

		go func() {
			errChan <- srv.ListenAndServe()
		}()
	}

	log.Fatal(<-errChan)
}
