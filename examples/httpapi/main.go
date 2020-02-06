package main

import (
	"fmt"
	"log"
	"net/http"

	httpimproc "github.com/ourstudio-se/go-image-processor/v2/http"
)

func main() {
	conv, err := httpimproc.NewHTTPImageConverter()
	if err != nil {
		panic(err)
	}

	api := &httpapi{
		conv,
	}

	http.Handle("/convert", api)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type httpapi struct {
	conv *httpimproc.HTTPImageConverter
}

func (h *httpapi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := h.conv.Read(r)
	if err != nil {
		log.Printf("could not handle request: %v\n", err)
		w.WriteHeader(400)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(b)))
	w.Header().Set("Content-Disposition", "inline")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
