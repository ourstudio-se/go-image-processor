package restful

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ourstudio-se/go-image-processor/abstractions"
	"github.com/ourstudio-se/go-image-processor/readers"
)

var defaultHTTPPort = 8080

// APIOptions enables customization for how the HTTP
// server is started
type APIOptions struct {
	Port     int
	Endpoint string
}

// API contains all references needed to convert
// a processing request
type API struct {
	opts          *APIOptions
	readerFactory *readers.URLReaderFactory
	converter     abstractions.Converter
}

// NewAPIOptions instantiates a new options
// container with default values
func NewAPIOptions(endpoint string) *APIOptions {
	return &APIOptions{
		Port:     defaultHTTPPort,
		Endpoint: endpoint,
	}
}

// NewAPI instantiates a new HTTP handler for converting images
func NewAPI(opts *APIOptions, readerFactory *readers.URLReaderFactory, converter abstractions.Converter) *API {
	return &API{
		opts,
		readerFactory,
		converter,
	}
}

// Start initiates a blocking HTTP listener
func (api *API) Start() error {
	router := httprouter.New()
	router.GET(api.opts.Endpoint, api.processingRequestHandler)

	port := fmt.Sprintf(":%d", api.opts.Port)
	return http.ListenAndServe(port, router)
}

func (api *API) processingRequestHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	preq, err := NewProcessingRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	rd := api.readerFactory.NewURLReader(preq.Source)

	if err = rd.ValidateURL(); err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(err.Error()))
		return
	}

	source, err := rd.ReadBlob()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	blob, err := api.converter.Apply(source, preq.OutputSpec)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(blob)))
	w.Header().Set("Content-Disposition", "inline")
	w.WriteHeader(http.StatusOK)
	w.Write(blob)
}
