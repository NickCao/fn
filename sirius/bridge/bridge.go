package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

var sock = flag.String("s", "vsock.sock", "path to unix socket")
var port = flag.Uint("p", 1, "port of vsock to connect")
var addr = flag.String("l", "127.0.0.1:2022", "address to listen on")

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) error {
	defer conn.Close()
	uconn, err := net.Dial("unix", *sock)
	if err != nil {
		return err
	}
	defer uconn.Close()
	_, err = fmt.Fprintf(uconn, "CONNECT %d\n", *port)
	if err != nil {
		return err
	}
	var ack uint
	_, err = fmt.Fscanf(uconn, "OK %d\n", &ack)
	if err != nil {
		return err
	}
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(conn, uconn)
	}()
	go func() {
		defer wg.Done()
		io.Copy(uconn, conn)
	}()
	wg.Wait()
	return nil
}
