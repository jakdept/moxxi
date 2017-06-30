package moxxi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

type rewriteProxy struct {
	client   *http.Client
	replacer Replacer
	down     string // the downstream domain
	up       string // the upstream domain
	port     int
	portTLS  int
	IP       net.IP
}

func (h *rewriteProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err, ok := recover().(error); ok && err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}()

	// does not currently handle other ports
	r.URL.Host = h.up
	childRequest, err := http.NewRequest(r.Method, r.URL.String(), h.replacer.RewriteRequest(r.Body))
	if err != nil {
		panic(err)
	}

	h.headerRewrite(&r.Header, &childRequest.Header, strings.NewReplacer(h.down, h.up))
	client := &http.Client{}
	resp, err := client.Do(childRequest)
	if err != nil {
		panic(err)
	}

	h.replacer.Reverse()
	writeHeader := w.Header()
	h.headerRewrite(&resp.Header, &writeHeader, strings.NewReplacer(h.up, h.down))
	h.replacer.RewriteResponse(resp.Body, w)
}

func (h *rewriteProxy) headerRewrite(in, out *http.Header, mod *strings.Replacer) {
	for headerName, headerSlice := range *in {
		for _, eachHeader := range headerSlice {
			out.Add(mod.Replace(headerName), mod.Replace(eachHeader))
		}
	}
}

// Replacer allows you to rewrite content, replacing old with new. It carries
// multiple methods to do replacement, but mostly works from an io.Reader to an
// io.Writer.
//
// THESE METHODS MAY PANIC.
type Replacer struct {
	old, new []byte
	memUsage int
}

// Reverse reverses old with the new in the Replacer.
func (h *Replacer) Reverse() {
	h.old, h.new = h.new, h.old
}

func (h *Replacer) replace(in io.Reader, out io.Writer) {
	if h.memUsage < 1 {
		h.memUsage = 1 << 20
	}
	if h.memUsage < 3*(len(h.old)+10) {
		panic(errors.New("not enough memory allocated for memory slice"))
	}

	type pair struct {
		l int
		r int
	}

	var err error
	var work pair
	var byteCount, next int

	data := make([]byte, h.memUsage)
	// cheat the first copy

	work.l = len(data)

	log.Printf("preflight l: %d r: %d n: %d data: %d/%d\n\n",
		work.l, work.r, next, len(data), cap(data))

	for {
		byteCount = copy(data, data[work.l:])
		work.r, err = in.Read(data[byteCount:])
		work.r += byteCount
		work.l = 0
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		// as long as there is enough room left to eat another chunk
		for work.l+len(h.old)+10 < work.r {
			fmt.Printf("writing loop - l: %d len: %d r: %d\n\n",
				work.l, len(h.old), work.r)
			if next = bytes.Index(data[work.l:work.r], h.old); next < 0 {
				next = work.r - len(h.old) - 10
				h.wholeWriter(out, data[work.l:next])
				work.l = next
				break
			}
			next += work.l
			h.wholeWriter(out, data[work.l:next])
			work.l = next + len(h.old)
			h.wholeWriter(out, h.new)
		}
	}

	fmt.Printf("l: %d r: %d remaining [%q]\n\n", work.l, work.r, data[work.l:work.r])

	h.wholeWriter(out, bytes.Replace(data[work.l:work.r], h.old, h.new, -1))
}

func (h *Replacer) wholeWriter(out io.Writer, data []byte) {
	fmt.Printf("writing - [%q]\n\n", data)
	var bc, w int
	var err error
	for w < len(data) {
		if bc, err = out.Write(data[w:]); err != nil {
			panic(err)
		}
		w += bc
	}
}

// RewriteRequest rewrites the request from old to new. It expects to run in a thread concurrently.
// If there are any errors, they will be returned via the error channel.
func (h *Replacer) RewriteRequest(in io.ReadCloser) io.ReadCloser {
	pipeReader, pipeWriter := io.Pipe()
	// make sure you close the input
	defer in.Close()
	go h.replace(in, pipeWriter)
	return pipeReader
}

func (h *Replacer) RewriteResponse(in io.ReadCloser, out io.Writer) {
	defer in.Close()
	go h.replace(in, out)
}
