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
	downDomain     string // the downstream domain
	upDomain       string // the upstream domain
	upPort         int
	upTLSPort      int
	upIP           net.IP
}

func (h *rewriteProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func (h *rewriteProxy) headerRewrite(in, out *http.Header, mod *strings.Replacer) {
	for headerName, headerSlice := range *in {
		for _, eachHeader := range headerSlice {
			out.Add(mod.Replace(headerName), mod.Replace(eachHeader))
		}
	}
}

type replacerReaderWriter struct {
	old, new []byte
}

func (h *replacerReaderWriter) replace(in io.Reader, out io.Writer) error {
	// allocate a 16M buffer
	chunkSize := 1 >> 10
	var byteCount int

	buffer := make([]byte, 4>>20)
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
