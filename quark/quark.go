package main

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/stream"
)

func MustLookupEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	} else {
		panic(fmt.Sprintf("environment variable %s is required", key))
	}
}

func EncodeTag(buf []byte) string {
	hash := sha256.Sum256(buf)
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash[:])
}

func main() {
	repo, err := name.NewRepository(MustLookupEnv("QUARK_REPO"))
	if err != nil {
		panic(err)
	}
	auth := remote.WithAuth(&authn.Basic{Username: MustLookupEnv("QUARK_USER"), Password: MustLookupEnv("QUARK_PASSWD")})

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		img, err := remote.Image(repo.Tag(EncodeTag([]byte(r.URL.Path))))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		layers, err := img.Layers()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(layers) != 1 {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data, err := layers[0].Uncompressed()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		io.Copy(w, data)
	})
	router.Put("/*", func(w http.ResponseWriter, r *http.Request) {
		layer := stream.NewLayer(r.Body)
		err := remote.WriteLayer(repo, layer, auth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		image, err := mutate.AppendLayers(empty.Image, layer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = remote.Write(repo.Tag(EncodeTag([]byte(r.URL.Path))), image, auth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.ListenAndServe(":3000", router)
}
