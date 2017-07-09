package moxxi

import (
	"net/http"
	"path"

	"github.com/jakdept/dandler"
)

func RefServerMuxer() http.Handler {
	funcList := []struct {
		p string
		h http.Handler
	}{
		{
			"/",
			dandler.Success("target: successful request"),
		}, {
			"/200",
			dandler.Success("target: successful request"),
		}, {
			"/301",
			http.RedirectHandler("/", http.StatusMovedPermanently),
		}, {
			"/302",
			http.RedirectHandler("/", http.StatusFound),
		}, {
			"/304",
			dandler.ResponseCode(http.StatusNotModified, "target: not modified"),
		}, {
			"/307",
			http.RedirectHandler("/", http.StatusTemporaryRedirect),
		}, {
			"/308",
			http.RedirectHandler("/", http.StatusPermanentRedirect),
		}, {
			"/401",
			dandler.ResponseCode(http.StatusUnauthorized, "target: authorization required"),
		}, {
			"/403",
			dandler.ResponseCode(http.StatusForbidden, "target: authorization required"),
		}, {
			"/404",
			dandler.ResponseCode(http.StatusNotFound, "target: not found"),
		}, {
			"/500",
			dandler.ResponseCode(http.StatusInternalServerError, "target: internal server error"),
		}, {
			"/503",
			dandler.ResponseCode(http.StatusServiceUnavailable, "target: gateway timeout"),
		},
	}

	mux := http.NewServeMux()
	for _, setup := range funcList {
		mux.Handle(path.Clean(setup.p), setup.h)
		mux.Handle(path.Clean(setup.p)+"/", setup.h)
	}
	h := dandler.Header("target", "this is a titled header", mux)
	h = dandler.Header("label", "the target is now in the body", h)
	return h
}

func RefServerListenAndServe(l string, h http.Handler) error {
	return http.ListenAndServe(l, h)
}
