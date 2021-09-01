package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("l", "[::]:8080", "listen address")
var path = flag.String("p", "./", "path to serve")

func main() {
	flag.Parse()
	log.Fatal(http.ListenAndServe(*addr, http.FileServer(http.Dir(*path))))
}
