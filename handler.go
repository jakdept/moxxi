package moxxi

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type rewriteProxy struct {
	client   *http.Client
	replacer Replacer
	down     string // the downstream domain
	up       string // the upstream domain
	port     int
	IP       net.IP
}

func (h *rewriteProxy) setup() error {
	dialer, err := StaticDialContext(h.IP, h.port)
	if err != nil {
		return err
	}

	h.client = &http.Client{
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           dialer,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	return nil
}

func (h *rewriteProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err, ok := recover().(error); ok && err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}()

	if h.client == nil {
		if err := h.setup(); err != nil {
			panic(err)
		}
	}

	fmt.Println(r)
	fmt.Printf("proxy host %s - url %s\n\n", r.Host, r.URL.String())

	replace := strings.NewReplacer(h.down, h.up)

	// does not currently handle other ports
	childRequest := &http.Request{
		Method:     r.Method,
		URL:        r.URL,
		Proto:      r.Proto,
		ProtoMajor: r.ProtoMajor,
		ProtoMinor: r.ProtoMinor,
		Header:     *h.headerRewrite(&r.Header, replace),
		Body:       h.replacer.RewriteRequest(r.Body),
		Close:      false,
	}

	if childRequest.URL.Host == "" {
		childRequest.URL.Host = r.Host
	}
	host, portStr, err := net.SplitHostPort(childRequest.URL.Host)
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}

	if childRequest.URL.Scheme != "https" &&
		childRequest.URL.Scheme != "http" &&
		port == 443 {
		childRequest.URL.Scheme = "https"
	} else if childRequest.URL.Scheme != "https" &&
		childRequest.URL.Scheme != "http" {
		childRequest.URL.Scheme = "http"
	}

	host = replace.Replace(host)
	if h.port != 0 {
		port = h.port
	}

	childRequest.URL.Host = host + ":" + strconv.Itoa(port)

	fmt.Printf("%#v\n\n", *childRequest.URL)
	fmt.Printf("%#v\n\n", childRequest)
	fmt.Printf("child host %s - url %s\n\n", r.Host, r.URL.String())

	dialer, err := StaticDialContext(h.IP, h.port)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           dialer,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	resp, err := client.Do(childRequest)
	if err != nil {
		panic(err)
	}

	h.replacer.Reverse()
	replace = strings.NewReplacer(h.up, h.down)
	parentHeader := w.Header()
	h.headerCopy(h.headerRewrite(&resp.Header, replace), &parentHeader)
	// h.headerRewrite(&resp.Header, &parentHeader, replace)
	done := h.replacer.Replace(resp.Body, w)
	<-done
}

func (h *rewriteProxy) headerCopy(in, out *http.Header) {
	for headerName, headerSlice := range *in {
		for _, eachHeader := range headerSlice {
			out.Add(headerName, eachHeader)
		}
	}
}

func (h *rewriteProxy) headerRewrite(in *http.Header, mod *strings.Replacer) *http.Header {
	out := &http.Header{}
	for headerName, headerSlice := range *in {
		for _, eachHeader := range headerSlice {
			out.Add(mod.Replace(headerName), mod.Replace(eachHeader))
		}
	}
	return out
}
