package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/JackKnifed/moxxi/moxxiconf"
	gorillaHandlers "github.com/gorilla/handlers"
)

func main() {
	listens, accessLogFile, errorLogFile, handlers, err := moxxiConf.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	var logger *log.Logger
	var err error
	var errorLog, accessLog *io.Writer

	if errorLogFile != "" {
		errorLog, err = os.OpenFile(errorLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			errorLog = os.Stderr
		}
	}

	if accessLogFile != "" {
		accessLog, err = os.OpenFile(accessLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			errorLog = os.Stdout
		}
	}

	logger := log.New(os.Stderr, "", log.LstdFlags|log.LUTC|log.Lshortfile)
	mux := moxxiConf.CreateMux(handlers, logger)

	var errChan chan error

	for _, singleListener := range listens {
		srv := http.Server{
			Addr:         singleListener,
			Handler:      gorillaHandlers.LoggingHandler(accessLog, mux),
			ReadTimeout:  moxxiConf.ConnTimeout,
			WriteTimeout: moxxiConf.ConnTimeout,
		}

		go func() {
			errChan <- srv.ListenAndServe()
		}()
	}

	log.Fatal(<-errChan)
}
