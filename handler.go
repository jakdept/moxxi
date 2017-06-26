package moxxi

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"strings"
)

type rewriteProxy struct {
	strConvertUp   *strings.Replacer
	strConvertDown *strings.Replacer
	client         *http.Client
	replacer       Replacer
	downDomain     string // the downstream domain
	upDomain       string // the upstream domain
	upPort         int
	upTLSPort      int
	upIP           net.IP
}

func (h *rewriteProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var childHeader http.Header
	h.headerRewrite(&r.Header, &childHeader, strings.NewReplacer(h.downDomain, h.upDomain))
	var errChan chan error

	childRequest := http.Request{
		Method: r.Method,
		URL:    r.URL,
		Header: childHeader,
		Body:   h.replacer.RewriteRequest(r.Body, errChan),
		Host:   h.upDomain,
	}
	client := &http.Client{}
	_, _ = client, childRequest

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
type Replacer struct {
	old, new []byte
}

// Reverse reverses old with the new in the Replacer.
func (h *Replacer) Reverse() {
	h.old, h.new = h.new, h.old
}

func (h *Replacer) replace(in io.Reader, out io.Writer) error {
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
				return err
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
			return err
		}
		buffer = buffer[byteCount:]
	}

	for len(buffer) > 0 {
		byteCount, err = out.Write(buffer)
		if err != nil {
			return err
		}
		buffer = buffer[byteCount:]
	}
	return nil
}

// RewriteRequest rewrites the request from old to new. It expects to run in a thread concurrently.
// If there are any errors, they will be returned via the error channel.
func (h *Replacer) RewriteRequest(in io.ReadCloser, errChan chan<- error) io.ReadCloser {
	pipeReader, pipeWriter := io.Pipe()
	// make sure you close the input
	defer in.Close()
	go func() {
		err := h.replace(in, pipeWriter)
		if err != nil {
			errChan <- err
		}
	}()
	return pipeReader
}

func (h *Replacer) RewriteResponse(in io.ReadCloser, out io.Writer, errChan chan<- error) {
	defer in.Close()
	go func() {
		err := h.replace(in, out)
		if err != nil {
			errChan <- err
		}
	}()
}
