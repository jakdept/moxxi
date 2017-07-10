package refSvr

import (
	"net/http"
	"path"

	"github.com/jakdept/dandler"
)

func BuildMuxer() *http.ServeMux {
	funcList := []struct {
		p string
		h http.Handler
	}{
		{
			"/",
			dandler.Success("refSvr: successful request"),
		}, {
			"/200",
			dandler.Success("refSvr: successful request"),
		}, {
			"/301",
			http.RedirectHandler("/", http.StatusMovedPermanently),
		}, {
			"/302",
			http.RedirectHandler("/", http.StatusFound),
		}, {
			"/304",
			dandler.ResponseCode(http.StatusNotModified, "refSvr: not modified"),
		}, {
			"/307",
			http.RedirectHandler("/", http.StatusTemporaryRedirect),
		}, {
			"/308",
			http.RedirectHandler("/", http.StatusPermanentRedirect),
		}, {
			"/401",
			dandler.ResponseCode(http.StatusUnauthorized, "refSvr: authorization required"),
		}, {
			"/403",
			dandler.ResponseCode(http.StatusForbidden, "refSvr: authorization required"),
		}, {
			"/404",
			dandler.ResponseCode(http.StatusNotFound, "refSvr: not found"),
		}, {
			"/500",
			dandler.ResponseCode(http.StatusInternalServerError, "refSvr: internal server error"),
		}, {
			"/503",
			dandler.ResponseCode(http.StatusServiceUnavailable, "refSvr: gateway timeout"),
		},
	}

	mux := http.NewServeMux()
	for _, setup := range funcList {
		mux.Handle(path.Clean(setup.p), setup.h)
		mux.Handle(path.Clean(setup.p)+"/", setup.h)
	}
	return mux
}

func ListenAndServe(l string, h http.Handler) error {
	h = dandler.Header("server", "testserver", h)
	h = dandler.Header("refSvr", "this is a titled header", h)
	h = dandler.Header("label", "the refSvr is now in the body", h)
	return http.ListenAndServe(l, h)
}
