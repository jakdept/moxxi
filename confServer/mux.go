package moxxiConf

import (
	"flag"
	"net/http"
)

func main() {
	var jsonHandler *HandlerLocFlag
	var formHandler *HandlerLocFlag
	var fileHandler *HandlerLocFlag
	var fileDocroot *HandlerLocFlag

	listen := flag.String("listen", ":8080", "listen address to use")
	confTempl := flag.String("confTempl", "template.conf", "base templates for the configs")
	resTempl := flag.String("resTempl", "template.response", "base template for the response")
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

	mux := http.NewServeMux()

	

	for _, each := range *jsonHandler {
		mux.HandleFunc(each, JSONHandler(baseDomain, confLoc, confExt, confTempl, resTempl, randHost))
	}

	for _, each := range *formHandler {
		mux.HandleFunc(each, JSONHandler(baseDomain, confLoc, confExt, confTempl, resTempl, randHost))
	}

	for i := 0; i < len(*fileHandler); i++ {
		mux.HandleFunc(*fileHandler[i], http.FileServer(*fileDocroot[i]))
	}

}
