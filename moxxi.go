package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/JackKnifed/moxxi/moxxiconf"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/natefinch/lumberjack"
)

// BroadCastSignal() a single signal across all channels ain an array
func BroadcastSignal(in chan os.Signal, out []chan os.Signal, done chan struct{}) {
	var val os.Signal
	for {
		select {
		case val = <-in:
			fmt.Println("got a usr1")
			for _, each := range out {
				each <- val
			}
		case <-done:
			return
		}
	}
}

func RotateLog(l *lumberjack.Logger, c chan os.Signal, done chan struct{}) {
	for {
		select {
		case <-c:
			l.Rotate()
		case <-done:
			return
		}
	}
}

func main() {
	var err error

	listens, accessLogFile, errorLogFile, handlers, err := moxxiConf.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("errorLog - %s - accessLog - %s", errorLogFile, accessLogFile)
	sigUsr := make(chan os.Signal, 1)

	done := make(chan struct{})
	defer close(done)

	signal.Notify(sigUsr, syscall.SIGUSR1)
	sigArr := []chan os.Signal{}

	var errorLog, accessLog io.Writer
	if errorLogFile != "" {
		intLogger := &lumberjack.Logger{
			Filename:   errorLogFile,
			MaxBackups: 5,
		}
		myChan := make(chan os.Signal)

		go RotateLog(intLogger, myChan, done)
		sigArr = append(sigArr, myChan)
		errorLog = intLogger
	} else {
		errorLog = os.Stderr
	}

	if accessLogFile != "" {
		intLogger := &lumberjack.Logger{
			Filename:   accessLogFile,
			MaxBackups: 5,
		}
		myChan := make(chan os.Signal)

		go RotateLog(intLogger, myChan, done)
		sigArr = append(sigArr, myChan)
		accessLog = intLogger
	} else {
		errorLog = os.Stdout
	}

	fmt.Printf("errorLog - %#v - accessLog - %#v", errorLog, accessLog)

	go BroadcastSignal(sigUsr, sigArr, done)

	logger := log.New(errorLog, "", log.LstdFlags|log.LUTC|log.Lshortfile)
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
