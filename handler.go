package moxxi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
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

	fmt.Println(r)
	fmt.Printf("proxy host %s - url %s\n\n", r.Host, r.URL.String())

	replace := strings.NewReplacer(h.down, h.up)

	// does not currently handle other ports
	childRequest, err := http.NewRequest(r.Method, replace.Replace(r.URL.String()),
		h.replacer.RewriteRequest(r.Body))
	if err != nil {
		panic(err)
	}

	h.headerRewrite(&r.Header, &childRequest.Header, replace)
	fmt.Printf("%#v\n", childRequest)
	fmt.Printf("child host %s - url %s\n\n", r.Host, r.URL.String())

	// dialer := &net.Dialer{
	// 	KeepAlive: 30 * time.Second,
	// 	DualStack: true,
	// }

	// func (d *dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, err) {
	// 			addr = fmt.Sprintf("%s:%s", h.IP,  h.port)
	// 		return dialer.DialContext(ctx, network, addr)
	// }

	// client := &http.Client{
	// 	Transport: http.Transport{
	// 		   DialContext: (&net.Dialer{
	//       KeepAlive: 30 * time.Second,
	//       DualStack: true,
	// 	},).dialContext,
	// }

	// dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
	// 	addr = fmt.Sprintf("%s:%d", h.IP, h.port)
	// 	return (&net.Dialer{}).DialContext(ctx, network, addr)
	// }

	client := &http.Client{
		Transport: &http.Transport{
			// DialContext: dialContext,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				addr = fmt.Sprintf("%s:%d", h.IP, h.port)
				return (&net.Dialer{}).DialContext(ctx, network, addr)
			},
			// DialContext: (
			// 	&net.Dialer{
			//     Timeout:   30 * time.Second,
			//     KeepAlive: 30 * time.Second,
			//     DualStack: true,
			// }).DialContext = dialContext,
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
	h.headerRewrite(&resp.Header, &parentHeader, replace)
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

	// the maximum size (in bytes) of a single utf8 char allowing a bit of grace
	bit := 10

	data := make([]byte, h.memUsage)
	// cheat the first copy

	work.l = len(data)
	work.r = len(data)

	for {
		byteCount = copy(data, data[work.l:work.r])

		// my problem is that i'm not bounding the copy slice below
		// if i drop work.r into it, i might have a shrinking buffer on short reads?
		work.r, err = in.Read(data[byteCount:])
		work.r += byteCount

		work.l = 0
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		// as long as there is enough room left to eat another chunk
		for work.l+len(h.old)+bit < work.r {
			// if there is no next chunk, write out everything but one full replacement
			if next = bytes.Index(data[work.l:work.r], h.old); next < 0 {
				next = work.r - len(h.old) - bit
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

	h.wholeWriter(out, bytes.Replace(data[work.l:work.r], h.old, h.new, -1))
}

func (h *Replacer) wholeWriter(out io.Writer, data []byte) {
	var bc, w int
	var err error
	for w < len(data) {
		if bc, err = out.Write(data[w:]); err != nil {
			panic(err)
		}
		w += bc
	}
}

// RewriteRequest rewrites the request from old to new.
// It expects to run in a thread concurrently.
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
