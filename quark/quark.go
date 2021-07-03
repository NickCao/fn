package main

import (
	"flag"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"io"
	"log"
	"os"
)

var repoStr = flag.String("repo", "", "image repo")
var tag = flag.String("tag", "latest", "image tag")
var pull = flag.Bool("pull", false, "pull instead of push")
var user = flag.String("user", "", "username")
var pass = flag.String("pass", "", "password")

func main() {
	flag.Parse()
	repo, err := name.NewRepository(*repoStr)
	if err != nil {
		log.Fatal(err)
	}
	auth := remote.WithAuth(&authn.Basic{Username: *user, Password: *pass})

	if !*pull {
		layer := stream.NewLayer(os.Stdin)
		err = remote.WriteLayer(repo, layer, auth)
		if err != nil {
			log.Fatal(err)
		}
		image, err := mutate.AppendLayers(empty.Image, layer)
		if err != nil {
			log.Fatal(err)
		}
		err = remote.Write(repo.Tag(*tag), image, auth)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		img, err := remote.Image(repo.Tag(*tag))
		if err != nil {
			log.Fatal(err)
		}
		layers, err := img.Layers()
		if err != nil {
			log.Fatal(err)
		}
		data, err := layers[0].Uncompressed()
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(os.Stdout, data)
	}
}
