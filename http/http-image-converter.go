package httpimproc

import (
	"fmt"
	"net/http"

	improc "github.com/ourstudio-se/go-image-processor"
)

// HTTPImageConverter wraps ImageConverter and handles
// HTTP requests for applying formats to an image
type HTTPImageConverter struct {
	Converter *improc.ImageConverter
}

// NewHTTPImageConverter instantiates a new ImageConverter
// which is able to parse HTTP requests and process
// images accordingly
func NewHTTPImageConverter() *HTTPImageConverter {
	return &HTTPImageConverter{
		Converter: improc.NewImageConverter(),
	}
}

// Read is the handler function for a HTTP request, and
// returns the raw image blob after the image has been
// processed
func (hic *HTTPImageConverter) Read(r *http.Request) ([]byte, error) {
	hic.Converter.Tracer(fmt.Sprintf("go-image-processor-http: parsing URL %s", r.URL.String()))

	preq, err := ParseURL(r.URL)
	if err != nil {
		hic.Converter.Tracer("go-image-processor-http: parsing URL failed!")
		return nil, err
	}

	hic.Converter.Tracer("go-image-processor-http: reading requested URL")

	reader := NewURLReader(preq.Source)
	b, err := reader.ReadBlob()
	if err != nil {
		hic.Converter.Tracer("go-image-processor-http: failed reading requested URL!")
		return nil, err
	}

	hic.Converter.Tracer("go-image-processor-http: applying parsed output specification to source")

	output, err := hic.Converter.Apply(b, preq.OutputSpec)
	if err != nil {
		hic.Converter.Tracer("go-image-processor-http: failed applying output specification!")
		return nil, err
	}

	hic.Converter.Tracer("go-image-processor-http: returning processed byte blob")
	return output, nil
}
