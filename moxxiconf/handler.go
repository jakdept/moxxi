package moxxiConf

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

func CreateMux(handlers []HandlerConfig) *http.ServeMux {
	mux := http.NewServeMux()
	for _, handler := range handlers {
		switch handler.handlerType {
		case "json":
			mux.HandleFunc(handler.handlerRoute, ChunkedJSONHandler(handler))
		case "form":
			mux.HandleFunc(handler.handlerRoute, FormHandler(handler))
		case "static":
			mux.HandleFunc(handler.handlerRoute, StaticHandler(handler))
		}
	}
	log.Printf("%#v", mux)
	return mux
}

// FormHandler - creates and returns a Handler for both Query and Form requests
func FormHandler(config HandlerConfig) http.HandlerFunc {
	confWriter := confWrite(config)

	return func(w http.ResponseWriter, r *http.Request) {

		if extErr := r.ParseForm(); extErr != nil {
			http.Error(w, extErr.Error(), http.StatusBadRequest)
			return
		}

		if r.Form.Get("host") == "" {
			pkgErr := &NewErr{Code: ErrNoHostname}
			http.Error(w, pkgErr.Error(), http.StatusPreconditionFailed)
			log.Println(pkgErr.LogError(r))
			return
		}
		host := r.Form.Get("host")

		if r.Form.Get("ip") == "" {
			pkgErr := &NewErr{Code: ErrNoIP}
			http.Error(w, pkgErr.Error(), http.StatusPreconditionFailed)
			log.Println(pkgErr.LogError(r))
			return
		}

		tls := parseCheckbox(r.Form.Get("tls"))

		port, err := strconv.Atoi(r.Form.Get("port"))
		if err != nil {
			port = 80
		}

		vhost := siteParams{
			IntHost:      host,
			IntIP:        r.Form.Get("ip"),
			Encrypted:    tls,
			IntPort:      port,
			StripHeaders: r.Form["header"],
		}

		vhost, pkgErr := confCheck(vhost, config)
		if pkgErr != nil {
			http.Error(w, pkgErr.Error(), http.StatusPreconditionFailed)
			log.Println(pkgErr.LogError(r))
			return
		}

		if vhost, pkgErr = confWriter(vhost); pkgErr != nil {
			http.Error(w, pkgErr.Error(), http.StatusInternalServerError)
			log.Println(pkgErr.LogError(r))
			return
		}

		if extErr := config.resTempl.Execute(w, []siteParams{vhost}); extErr != nil {
			http.Error(w, pkgErr.Error(), http.StatusInternalServerError)
			log.Println(pkgErr.LogError(r))
			return
		}
		return
	}
}

// JSONHandler - creates and returns a Handler for JSON body requests
func ChunkedJSONHandler(config HandlerConfig) http.HandlerFunc {

	var tStart, tEnd, tBody, tError *template.Template

	for _, each := range config.resTempl.Templates() {
		switch each.Name() {
		case "start":
			tStart = each
		case "end":
			tEnd = each
		case "body":
			tBody = each
		case "error":
			tError = each
		}
	}

	if tStart == nil || tEnd == nil || tBody == nil || tError == nil {
		return JSONHandler(config)
	}

	confWriter := confWrite(config)

	return func(w http.ResponseWriter, r *http.Request) {

		var emptyInterface interface{}
		tStart.Execute(w, emptyInterface)

		decoder := json.NewDecoder(r.Body)
		var v struct {
			host           string
			ip             string
			port           int
			tls            bool
			blockedHeaders []string
		}

		for decoder.More() {
			err := decoder.Decode(&v)
			if err != nil {
				continue
			}

			vhost := siteParams{
				IntHost:      v.host,
				IntIP:        v.ip,
				Encrypted:    v.tls,
				IntPort:      v.port,
				StripHeaders: v.blockedHeaders,
			}

			confConfig, err := confCheck(vhost, config)
			if err != nil {
				tError.Execute(w, vhost)
			}

			// #TODO# maybe change the style of siteParams to include myError?
			if confConfig, err := confWriter(confConfig); err != nil {
				log.Println(err.LogError(r))
				erredConfig := struct {
					ExtHost      string
					IntHost      string
					IntIP        string
					IntPort      int
					Encrypted    bool
					StripHeaders []string
					myError      string
				}{
					ExtHost:      confConfig.ExtHost,
					IntHost:      confConfig.IntHost,
					IntIP:        confConfig.IntIP,
					IntPort:      confConfig.IntPort,
					Encrypted:    confConfig.Encrypted,
					StripHeaders: confConfig.StripHeaders,
					myError:      err.Error(),
				}
				if err := tError.Execute(w, erredConfig); err != nil {
					log.Println(err.Error())
				}
			}

			if err = tBody.Execute(w, confConfig); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println(err.Error())
				return
			}
		}
		tEnd.Execute(w, emptyInterface)
	}
}

func JSONHandler(config HandlerConfig) http.HandlerFunc {
	confWriter := confWrite(config)
	return func(w http.ResponseWriter, r *http.Request) {
		var v []struct {
			host           string
			ip             string
			port           int
			tls            bool
			blockedHeaders []string
		}
		var responseConfig []siteParams

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		for _, each := range v {

			vhost := siteParams{
				IntHost:      each.host,
				IntIP:        each.ip,
				Encrypted:    each.tls,
				IntPort:      each.port,
				StripHeaders: each.blockedHeaders,
			}

			confConfig, err := confCheck(vhost, config)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			confConfig, err = confCheck(vhost, config)
			if err != nil {
				http.Error(w, err.Error(), http.StatusPreconditionFailed)
				log.Println(err.LogError(r))
				return
			}

			if confConfig, err = confWriter(confConfig); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println(err.LogError(r))
				return
			}

			responseConfig = append(responseConfig, confConfig)
		}

		if err := config.resTempl.Execute(w, responseConfig); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}
	}
}

// StaticHandler - creates and returns a Handler to simply respond with a static response to every request
func StaticHandler(config HandlerConfig) http.HandlerFunc {
	res, err := ioutil.ReadFile(config.resFile)
	if err != nil {
		log.Printf("bad static response file %s - %v", config.resFile, err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}
