package main

import (
	"log"

	"github.com/jakdept/moxxi/refSvr"
)

func main() {
	h := refSvr.BuildMuxer()
	log.Fatal(refSvr.ListenAndServe(":8081", h))
}
