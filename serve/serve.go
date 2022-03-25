package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"
)

var addr = flag.String("l", "[::]:8080", "listen address")
var path = flag.String("p", "./", "path to serve")

func main() {
	flag.Parse()
	fs := http.FileServer(&FileSystem{
		FileSystem: http.Dir(*path),
		mod:        time.Now(),
	})
	log.Fatal(http.ListenAndServe(*addr, fs))
}

type FileSystem struct {
	http.FileSystem
	mod time.Time
}

func (f *FileSystem) Open(name string) (http.File, error) {
	file, err := f.FileSystem.Open(name)
	return &File{File: file, mod: f.mod}, err
}

type File struct {
	http.File
	mod time.Time
}

func (f *File) Stat() (os.FileInfo, error) {
	fileInfo, err := f.File.Stat()
	return &FileInfo{FileInfo: fileInfo, mod: f.mod}, err
}

type FileInfo struct {
	os.FileInfo
	mod time.Time
}

func (f *FileInfo) ModTime() time.Time {
	return f.mod
}
