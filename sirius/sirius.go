package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"flag"
	"github.com/mdlayher/vsock"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"os/exec"
	"sync"
)

var port = flag.Uint("p", 1, "vsock port to listen on")
var nix = flag.String("n", "nix-store", "path to nix-store command")

func main() {
	flag.Parse()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		log.Fatal(err)
	}

	var config = &ssh.ServerConfig{NoClientAuth: true}
	config.AddHostKey(signer)

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
		go handleConn(conn, config)
	}
}

func handleConn(conn net.Conn, config *ssh.ServerConfig) error {
	sconn, chrs, greqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return err
	}
	defer sconn.Close()

	wg := &sync.WaitGroup{}
	go ssh.DiscardRequests(greqs)
	for chr := range chrs {
		wg.Add(1)
		go handleChannelRequest(chr, wg)
	}
	wg.Wait()

	return nil
}

func handleChannelRequest(chr ssh.NewChannel, wg *sync.WaitGroup) error {
	defer wg.Done()

	if chr.ChannelType() != "session" {
		return chr.Reject(ssh.UnknownChannelType, "unknown channel type")
	}

	ch, reqs, err := chr.Accept()
	if err != nil {
		return err
	}

	for req := range reqs {
		if req.Type != "exec" {
			if req.WantReply {
				req.Reply(false, nil)
			}
			continue
		}
		if len(req.Payload) < 4 || !bytes.Equal(req.Payload[4:], []byte("nix-store --serve --write")) {
			if req.WantReply {
				req.Reply(false, nil)
			}
			continue
		}
		cmd := exec.Command(*nix, "--serve", "--write")
		cmd.Stdin = ch
		cmd.Stdout = ch
		cmd.Stderr = ch.Stderr()
		err = cmd.Start()
		if req.WantReply {
			req.Reply(err == nil, nil)
		}
	}
	return nil
}
