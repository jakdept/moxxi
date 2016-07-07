package moxxiConf

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

func CreateMux(handlers []HandlerConfig, l *log.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	for _, handler := range handlers {
		switch handler.handlerType {
		case "json":
			mux.HandleFunc(handler.handlerRoute, JSONHandler(handler, l))
		case "form":
			mux.HandleFunc(handler.handlerRoute, FormHandler(handler, l))
		case "static":
			mux.HandleFunc(handler.handlerRoute, StaticHandler(handler, l))
		}
	}
	return mux
}

// FormHandler - creates and returns a Handler for both Query and Form requests
func FormHandler(config HandlerConfig, l *log.Logger) http.HandlerFunc {
	confWriter := confWrite(config)

	return func(w http.ResponseWriter, r *http.Request) {

		if extErr := r.ParseForm(); extErr != nil {
			http.Error(w, extErr.Error(), http.StatusBadRequest)
			return
		}

		if r.Form.Get("host") == "" {
			pkgErr := &NewErr{Code: ErrNoHostname}
			http.Error(w, pkgErr.Error(), http.StatusPreconditionFailed)
			l.Println(pkgErr.LogError(r))
			return
		}
		host := r.Form.Get("host")

		if r.Form.Get("ip") == "" {
			pkgErr := &NewErr{Code: ErrNoIP}
			http.Error(w, pkgErr.Error(), http.StatusPreconditionFailed)
			l.Println(pkgErr.LogError(r))
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
			l.Println(pkgErr.LogError(r))
			return
		}

		if vhost, pkgErr = confWriter(vhost); pkgErr != nil {
			http.Error(w, pkgErr.Error(), http.StatusInternalServerError)
			l.Println(pkgErr.LogError(r))
			return
		}

		if extErr := config.resTempl.Execute(w, []siteParams{vhost}); extErr != nil {
			http.Error(w, extErr.Error(), http.StatusInternalServerError)
			l.Println(extErr.Error())
			return
		}
		return
	}
}

// JSONHandler - creates and returns a Handler for JSON body requests
func JSONHandler(config HandlerConfig, l *log.Logger) http.HandlerFunc {

	var tStart, tEnd, tBody *template.Template

	for _, each := range config.resTempl.Templates() {
		switch each.Name() {
		case "start":
			tStart = each
		case "end":
			tEnd = each
		case "body":
			tBody = each
		}
	}

	if tStart == nil || tEnd == nil || tBody == nil {
		return InvalidHandler("bad template", http.StatusInternalServerError)
	}

	confWriter := confWrite(config)

	return func(w http.ResponseWriter, r *http.Request) {

		var emptyInterface interface{}
		type locSiteParams struct {
			ExtHost      string
			IntHost      string
			IntIP        string
			IntPort      int
			Encrypted    bool
			StripHeaders []string
			ErrorString  string
			Error        Err
		}

		tStart.Execute(w, emptyInterface)

		decoder := json.NewDecoder(r.Body)

		for decoder.More() {
			var v siteParams
			if err := decoder.Decode(&v); err != nil {
				continue
			}

			v, err := confCheck(v, config)
			if err == nil {
				v, err = confWriter(v)
			} else if err.GetCode() == ErrBadHostnameTrace {
				var newErr Error
				v, newErr = confWriter(v)
				if newErr != nil {
					err = newErr
				}
			}

			var vPlus = struct {
				ExtHost      string
				IntHost      string
				IntIP        string
				IntPort      int
				Encrypted    bool
				StripHeaders []string
				Error        string
			}{
				ExtHost:      v.ExtHost,
				IntHost:      v.IntHost,
				IntIP:        v.IntIP,
				IntPort:      v.IntPort,
				Encrypted:    v.Encrypted,
				StripHeaders: v.StripHeaders,
			}

			if err != nil {
				l.Println(err.LogError(r))
				vPlus.Error = err.Error()
			}

			if err := tBody.Execute(w, vPlus); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				l.Println(err.Error())
				return
			}
		}
		tEnd.Execute(w, emptyInterface)
	}
}

// StaticHandler - creates and returns a Handler to simply respond with a static response to every request
func StaticHandler(config HandlerConfig, l *log.Logger) http.HandlerFunc {
	res, err := ioutil.ReadFile(config.resFile)
	if err != nil {
		l.Printf("bad static response file %s - %v", config.resFile, err)
		return InvalidHandler("no data", http.StatusInternalServerError)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func InvalidHandler(msg string, code int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, msg, code)
	}
}
