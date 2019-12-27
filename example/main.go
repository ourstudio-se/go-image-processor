package main

import (
	"fmt"
	"log"
	"net/http"

	httpimproc "github.com/ourstudio-se/go-image-processor/http"
)

type httpapi struct {
	conv *httpimproc.HTTPImageConverter
}

func (ha *httpapi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := ha.conv.Read(r)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(b)))
	w.Header().Set("Content-Disposition", "inline")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func main() {
	conv := httpimproc.NewHTTPImageConverter()
	api := &httpapi{
		conv,
	}

	http.Handle("/convert", api)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
