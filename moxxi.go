package main

import (
	"github.com/JackKnifed/moxxi/moxxiconf"
	gorillaHandlers "github.com/gorilla/handlers"
	"log"
	"net/http"
	"os"
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
