package main

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
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
	"strings"
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

const respOk = "<_/>"
const respErr = "<Error><Code>NoSuchKey</Code></Error>"

func main() {
	repo, err := name.NewRepository(MustLookupEnv("QUARK_REPO"))
	if err != nil {
		panic(err)
	}
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.CleanPath)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := strings.Split(strings.TrimPrefix(r.Header.Get("authorization"), "AWS4-HMAC-SHA256 "), " ")
			var auth authn.Authenticator = authn.Anonymous
			for _, part := range authorization {
				if strings.HasPrefix(part, "Credential=") {
					parts := strings.Split(strings.TrimPrefix(part, "Credential="), "/")
					if len(parts) != 0 {
						auth = &authn.Bearer{Token: parts[0]}
					}
				}
			}
			ctx := context.WithValue(r.Context(), "auth", remote.WithAuth(auth))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "quark: s3 compatible nix binary cache", http.StatusOK)
	})
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		img, err := remote.Image(repo.Tag(EncodeTag([]byte(chi.URLParam(r, "*")))))
		if err != nil {
			http.Error(w, respErr, http.StatusNotFound)
			return
		}
		layers, err := img.Layers()
		if err != nil {
			http.Error(w, respErr, http.StatusInternalServerError)
			return
		}
		if len(layers) != 1 {
			http.Error(w, respErr, http.StatusInternalServerError)
			return
		}
		data, err := layers[0].Uncompressed()
		if err != nil {
			http.Error(w, respErr, http.StatusInternalServerError)
			return
		}
		io.Copy(w, data)
	})
	router.Head("/{bucket}/*", func(w http.ResponseWriter, r *http.Request) {
		_, err := remote.Image(repo.Tag(EncodeTag([]byte(chi.URLParam(r, "*")))))
		if err != nil {
			http.Error(w, respErr, http.StatusNotFound)
			return
		}
		http.Error(w, respOk, http.StatusOK)
	})
	router.Put("/{bucket}/*", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Context().Value("auth").(remote.Option)
		layer := stream.NewLayer(r.Body)
		err := remote.WriteLayer(repo, layer, auth)
		if err != nil {
			http.Error(w, respErr, http.StatusInternalServerError)
			return
		}
		image, err := mutate.AppendLayers(empty.Image, layer)
		if err != nil {
			http.Error(w, respErr, http.StatusInternalServerError)
			return
		}
		err = remote.Write(repo.Tag(EncodeTag([]byte(chi.URLParam(r, "*")))), image, auth)
		if err != nil {
			http.Error(w, respErr, http.StatusInternalServerError)
			return
		}
		http.Error(w, respOk, http.StatusOK)
	})
	http.ListenAndServe(":3000", router)
}
