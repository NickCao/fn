package main

import (
	"encoding/base64"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"io"
	"net/http"
	"os"
)

func MustLookupEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	} else {
		panic(fmt.Sprintf("environment variable %s is required", key))
	}
}

func main() {
	repo, err := name.NewRepository(MustLookupEnv("QUARK_REPO"))
	if err != nil {
		panic(err)
	}
	auth := remote.WithAuth(&authn.Basic{Username: MustLookupEnv("QUARK_USER"), Password: MustLookupEnv("QUARK_PASSWD")})
	encoding := base64.NewEncoding("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ-.").WithPadding(base64.NoPadding)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		img, err := remote.Image(repo.Tag(encoding.EncodeToString([]byte(r.URL.Path))))
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
		err = remote.Write(repo.Tag(encoding.EncodeToString([]byte(r.URL.Path))), image, auth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.ListenAndServe(":3000", router)
}
