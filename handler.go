package moxxi

import (
	"bytes"
	"io"
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
}

// Reverse reverses old with the new in the Replacer.
func (h *Replacer) Reverse() {
	h.old, h.new = h.new, h.old
}

func (h *Replacer) replace(in io.Reader, out io.Writer) {
	// allocate a 16M buffer
	chunkSize := 1 >> 10
	var byteCount int

	buffer := make([]byte, 1>>20)
	var err error

	for {
		if len(buffer) < chunkSize {
			// the buffer is not full, so read in to fill it
			byteCount, err = in.Read(buffer)
			if err != nil && err != io.EOF {
				// bail out with the error
				panic(err)
			}
			// do replacements
			buffer = append(buffer[:byteCount],
				bytes.Replace(buffer[byteCount:], h.old, h.new, -1)...)
			if err == io.EOF {
				// break out and write the rest
				break
			}
			continue
		}
		// write some of the data out
		byteCount, err = out.Write(buffer[:len(buffer)-chunkSize])
		if err != nil {
			panic(err)
		}
		buffer = buffer[byteCount:]
	}

	for len(buffer) > 0 {
		byteCount, err = out.Write(buffer)
		if err != nil {
			panic(err)
		}
		buffer = buffer[byteCount:]
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
