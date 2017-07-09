package main

import "log"
import "github.com/jakdept/moxxi/support"

func main() {
	h := moxxi.RefServerMuxer()
	log.Fatal(moxxi.RefServerListenAndServe(":8081", h))
}
