package main

import (
	"flag"
	"github.com/mdlayher/vsock"
	"log"
	"net"
	"os"
	"os/exec"
)

var port = flag.Uint("p", 1, "vsock port to listen on")
var nix = flag.String("n", "nix-daemon", "path to nix-daemon command")

func main() {
	flag.Parse()

	lis, err := vsock.Listen(uint32(*port))
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %s\n", err)
			continue
		}
		log.Printf("accepted connection from: %s\n", conn.RemoteAddr())
		go func(c net.Conn) {
			err := handleConn(c)
			if err != nil {
				log.Printf("nix-daemon invocation error: %s\n", err)
			}
		}(conn)
	}
}

func handleConn(conn net.Conn) error {
	defer conn.Close()
	return (&exec.Cmd{
		Path:   *nix,
		Args:   []string{"nix-daemon", "--stdio"},
		Env:    []string{},
		Stdin:  conn,
		Stdout: conn,
		Stderr: os.Stderr,
	}).Run()
}
