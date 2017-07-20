package moxxi

import (
	"bytes"
	"errors"
	"io"
)

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

func (h *Replacer) Replace(in io.Reader, out io.Writer) <-chan struct{} {
	// check the required parameters
	if h.memUsage < 1 {
		h.memUsage = 1 << 20
	}
	if h.memUsage < 3*(len(h.old)+10) {
		panic(errors.New("not enough memory allocated for memory slice"))
	}

	// make and prime the done marker
	done := make(chan struct{})
	go func() {
		done <- struct{}{}
	}()
	<-done

	go h.replaceLoop(in, out, done)
	return done
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

func (h *Replacer) replaceLoop(in io.Reader, out io.Writer, done chan<- struct{}) {
	defer close(done)

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

// RewriteRequest rewrites the request from old to new.
// It expects to run in a thread concurrently.
// If there are any errors, they will be returned via the error channel.
func (h *Replacer) RewriteRequest(in io.ReadCloser) io.ReadCloser {
	pipeReader, pipeWriter := io.Pipe()
	// make sure you close the input
	// defer in.Close()
	done := h.Replace(in, pipeWriter)
	go func() {
		<-done
		pipeWriter.Close()
		// pipeReader.Close()
	}()
	return pipeReader
}
