package moxxiConf

import (
	"flag"
	"log"
	"net/http"
	"text/template"
)

func main() {
	var jsonHandler *HandlerLocFlag
	var formHandler *HandlerLocFlag
	var fileHandler *HandlerLocFlag
	var fileDocroot *HandlerLocFlag

	listen := flag.String("listen", ":8080", "listen address to use")
	confTemplString := flag.String("confTempl", "template.conf", "base templates for the configs")
	resTemplString := flag.String("resTempl", "template.response", "base template for the response")
	baseDomain := flag.String("domain", "", "base domain to add onto")
	subdomainLength := flag.Int("subLength", 8, "length of subdomain to exclude")
	excludedDomain := flag.String("excludedDomain", "", "domains to reject")
	confLoc := flag.String("confLoc", "", "path to put the domains")
	confExt := flag.String("confExt", ".conf", "extension to add to the confs")

	flag.Var(jsonHandler, "jsonHandler", "location for a JSON handler")
	flag.Var(formHandler, "formHandler", "location for a form handler")
	flag.Var(fileHandler, "fileHandler", "location for a file handler")
	flag.Var(fileDocroot, "fileDocroot", "docroots for each file handler")

	flag.Parse()

	confTempl, err := template.ParseFiles(*confTemplString)
	if err != nil {
		log.Fatal(err)
	}
	resTempl, err := template.ParseFiles(*resTemplString)
	if err != nil {
		log.Fatal(err)
	}

	var done chan struct{}
	mux := http.NewServeMux()

	randHost := randSeqFeeder(*baseDomain, *excludedDomain, *subdomainLength, done)

	for _, each := range *jsonHandler {
		mux.HandleFunc(each, JSONHandler(*baseDomain, *confLoc, *confExt, *confTempl, *resTempl, randHost))
	}

	for _, each := range *formHandler {
		mux.HandleFunc(each, JSONHandler(*baseDomain, *confLoc, *confExt, *confTempl, *resTempl, randHost))
	}

	for i := 0; i < len(*fileHandler); i++ {
		mux.Handle(fileHandler.GetOne(i), http.FileServer(http.Dir(fileDocroot.GetOne(i))))
	}

	srv := http.Server{
		Addr:         *listen,
		Handler:      mux,
		ReadTimeout:  ConnTimeout,
		WriteTimeout: ConnTimeout,
	}

	log.Fatal(srv.ListenAndServe())
}
