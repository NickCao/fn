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
			log.Println(err)
			continue
		}
		log.Printf("connection from: %s\n", conn.RemoteAddr().String())
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) error {
	return (&exec.Cmd{
		Path:   *nix,
		Args:   []string{"--stdio"},
		Env:    []string{},
		Stdin:  conn,
		Stdout: conn,
		Stderr: os.Stderr,
	}).Start()
}
