package moxxiConf

import (
	"encoding/json"
	"net/http"
	// "log"
	"strconv"
	"text/template"
)

// FormHandler - creates and returns a Handler for both Query and Form requests
func FormHandler(baseURL, confPath, confExt string, excludes []string,
	confTempl, resTempl template.Template, subdomainLen int) http.HandlerFunc {

	confWriter := confWrite(confPath, confExt, baseURL, subdomainLen, confTempl, excludes)

	return func(w http.ResponseWriter, r *http.Request) {

		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var tls bool

		if r.Form.Get("host") == "" {
			http.Error(w, "no provided hostname", http.StatusPreconditionFailed)
			// TODO some log line?
			return
		}

		if r.Form.Get("ip") == "" {
			http.Error(w, "no provided ip", http.StatusPreconditionFailed)
			// TODO some log line?
			return
		}

		if tls, err = strconv.ParseBool(r.Form.Get("tls")); err != nil {
			tls = DefaultBackendTLS
		}

		port, _ := strconv.Atoi(r.Form.Get("port"))
		config, err := confCheck(r.Form.Get("host"), r.Form.Get("ip"), tls, port,
			r.Form["header"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
			// TODO some log line?
			return
		}

		if config, err = confWriter(config); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			// TODO some log line? or no?
			return
		}

		if err = resTempl.Execute(w, []siteParams{config}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			// TODO some long line? or no?
			return
		}
		return
	}
}

// JSONHandler - creates and returns a Handler for JSON body requests
func JSONHandler(baseURL, confPath, confExt string, excludes []string,
	confTempl, resTempl template.Template, subdomainLen int) http.HandlerFunc {

	confWriter := confWrite(confPath, confExt, baseURL, subdomainLen, confTempl, excludes)

	return func(w http.ResponseWriter, r *http.Request) {

		var v []struct {
			host           string
			ip             string
			port           int
			tls            bool
			blockedHeaders []string
		}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		var responseConfig []siteParams

		for _, each := range v {
			config, err := confCheck(each.host, each.ip, each.tls, each.port, each.blockedHeaders)
			if err != nil {
				http.Error(w, err.Error(), http.StatusPreconditionFailed)
				// TODO some log line?
				return
			}

			if config, err = confWriter(config); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				// TODO some log line? or no?
				return
			}

			responseConfig = append(responseConfig, config)
		}

		if err = resTempl.Execute(w, responseConfig); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			// TODO some long line? or no?
			return
		}
		return
	}
}
