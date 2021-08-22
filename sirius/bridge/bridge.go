package main

import (
	"bytes"
	// "context"
	"crypto/ed25519"
	"crypto/rand"
	"flag"
	// "github.com/firecracker-microvm/firecracker-go-sdk"
	// "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	// "github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"sync"
)

var port = flag.Uint("p", 1024, "port of vsock to connect")
var addr = flag.String("l", "127.0.0.1:2022", "address to listen on")
var kernel = flag.String("k", "", "path to kernel image")
var kargs = flag.String("a", "", "kernel args")
var rootfs = flag.String("r", "", "path to rootfs")
var fc = flag.String("f", "", "path to firecracker binary")

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
		go func() {
			log.Println(handleConn(conn, config))
		}()
	}
}

func handleConn(conn net.Conn, config *ssh.ServerConfig) error {
	log.Println("new connection")
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
	log.Println("new channel")
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
		if len(req.Payload) < 4 || !bytes.Equal(req.Payload[4:], []byte("nix-daemon --stdio")) {
			if req.WantReply {
				req.Reply(false, nil)
			}
			continue
		}
		/*
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			base := "/tmp/" + uuid.NewString()
			t := true
			r := "rootfs"
			var c1024 int64 = 1024
			var c1 int64 = 1
			machine, err := firecracker.NewMachine(ctx, firecracker.Config{
				SocketPath:      base + ".ctrl",
				LogLevel:        "Info",
				KernelImagePath: *kernel,
				KernelArgs:      *kargs,
				Drives: []models.Drive{{
					DriveID:      &r,
					IsReadOnly:   &t,
					IsRootDevice: &t,
					PathOnHost:   rootfs,
				}},
				VsockDevices: []firecracker.VsockDevice{{
					ID:   "nix",
					Path: base + ".vsock",
					CID:  3,
				}},
				MachineCfg: models.MachineConfiguration{
					HtEnabled:  &t,
					MemSizeMib: &c1024,
					VcpuCount:  &c1,
				},
				VMID: uuid.NewString(),
			}, firecracker.WithProcessRunner(firecracker.VMCommandBuilder{}.
				WithBin(*fc).
				WithSocketPath(base+".ctrl").
				Build(ctx)))
			if err != nil {
				return err
			}
			err = machine.Start(ctx)
			if err != nil {
				return err
			}
			uconn, err := VSockDial(ctx, base+".vsock", uint32(*port))
			if err != nil {
				return err
			}
		*/
		if req.WantReply {
			req.Reply(true, nil)
		}
		go func() {
			err = (&Daemon{}).ProcessConn(ch)
			log.Println(err)
		}()
	}
	return nil
}
