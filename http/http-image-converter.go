package httpimproc

import (
	"net/http"

	improc "github.com/ourstudio-se/go-image-processor"
)

// HTTPImageConverter wraps ImageConverter and handles
// HTTP requests for applying formats to an image
type HTTPImageConverter struct {
	converter *improc.ImageConverter
}

// NewHTTPImageConverter instantiates a new ImageConverter
// which is able to parse HTTP requests and process
// images accordingly
func NewHTTPImageConverter() *HTTPImageConverter {
	return &HTTPImageConverter{
		converter: improc.NewImageConverter(),
	}
}

// Read is the handler function for a HTTP request, and
// returns the raw image blob after the image has been
// processed
func (hic *HTTPImageConverter) Read(r *http.Request) ([]byte, error) {
	preq, err := ParseURL(r.URL)
	if err != nil {
		return nil, err
	}

	reader := NewURLReader(preq.Source)
	b, err := reader.ReadBlob()
	if err != nil {
		return nil, err
	}

	output, err := hic.converter.Apply(b, preq.OutputSpec)
	if err != nil {
		return nil, err
	}

	return output, nil
}
