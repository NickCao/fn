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
	fs := http.FileServer(http.Dir(*path))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("If-Modified-Since")
		fs.ServeHTTP(w, r)
	})
	log.Fatal(http.ListenAndServe(*addr, handler))
}
