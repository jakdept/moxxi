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

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/moxxi/")
	viper.AddConfigPath("$HOME/.moxxi")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	var config moxxiconf.MoxxiConf
	err = Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode config into struct, %v", err)
	}

	err = moxxiconf.CheckConfig(&config)
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
