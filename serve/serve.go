package main

import (
	"flag"
	"net/http"
)

var addr = flag.String("l", "[::]:8080", "listen address")
var path = flag.String("p", "./", "path to serve")

func main() {
	http.ListenAndServe(*addr, http.FileServer(http.Dir(*path)))
}
