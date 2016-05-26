package moxxiConf

import (
	"encoding/json"
	"net/http"
	// "log"
	"strconv"
	"text/template"
)

// FormHandler - creates and returns a Handler for both Query and Form requests
func FormHandler(config) http.HandlerFunc {
	confWriter := confWrite(config)

	return func(w http.ResponseWriter, r *http.Request) {

		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var tls bool

		if r.Form.Get("host") == "" {
			err = &Err{Code: ErrNoHostname}
			http.Error(w, err, http.StatusPreconditionFailed)
			log.Println(err.LogError(r))
			return
		}

		if r.Form.Get("ip") == "" {
			err = &Err{Code: ErrNoIP}
			http.Error(w, err, http.StatusPreconditionFailed)
			log.Println(err.LogError(r))
			return
		}

		if tls = parseCheckbox(r.Form.Get("tls")); err != nil {
			tls = DefaultBackendTLS
		}

		port, _ := strconv.Atoi(r.Form.Get("port"))
		siteConfig, err := confCheck(r.Form.Get("host"), r.Form.Get("ip"), tls, port,
			r.Form["header"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
			log.Println(err.LogError(r))
			return
		}

		if siteConfig, err = confWriter(siteConfig); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err.LogError(r))
			return
		}

		if err = resTempl.Execute(w, []siteParams{siteConfig}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err.LogError(r))
			return
		}
		return
	}
}

// JSONHandler - creates and returns a Handler for JSON body requests
func JSONHandler(config HandlerConfig) http.HandlerFunc {

	confWriter := confWrite(config)

	return func(w http.ResponseWriter, r *http.Request) {

		// TODO move this stuff so it's declared once
		var v []struct {
			host           string
			ip             string
			port           int
			tls            bool
			blockedHeaders []string
		}

		decoder := json.NewDecoder(r.Body)
		// TODO this probably introduces a bug where only one json array is decoded
		err := decoder.Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		var responseConfig []siteParams

		for _, each := range v {
			confConfig, err := confCheck(each.host, each.ip, each.tls, each.port, each.blockedHeaders)
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

		if err = resTempl.Execute(w, responseConfig); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err.LogError(r))
			return
		}
		return
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
