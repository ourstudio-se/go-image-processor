package httpimproc

import (
	"fmt"
	"net/http"

	improc "github.com/ourstudio-se/go-image-processor/v2"
)

// ParameterMap is the definitions used
// when parsing a request, to extract the
// information needed to do image processing
type ParameterMap struct {
	SourceURL      string
	Width          string
	Height         string
	Crop           string
	Quality        string
	Spec           string
	AnchorX        string
	AnchorY        string
	Compression    string
	TextValue      string
	TextFont       string
	TextSize       string
	TextForeground string
	TextBackground string
	TextAnchor     string
	Background     string
}

// DefaultParameterMap returns a ParameterMap with
// the default querystring parameter names used
// when parsing a request
func DefaultParameterMap() *ParameterMap {
	return &ParameterMap{
		SourceURL:      "url",
		Width:          "width",
		Height:         "height",
		Crop:           "crop",
		Quality:        "quality",
		Spec:           "spec",
		AnchorX:        "anchorx",
		AnchorY:        "anchory",
		Compression:    "output",
		TextValue:      "text:value",
		TextFont:       "text:font",
		TextSize:       "text:size",
		TextForeground: "text:foreground",
		TextBackground: "text:background",
		TextAnchor:     "text:anchor",
		Background:     "background",
	}
}

// HTTPImageConverter wraps ImageConverter and handles
// HTTP requests for applying formats to an image
type HTTPImageConverter struct {
	Converter    *improc.ImageConverter
	ParemeterMap *ParameterMap
}

// NewHTTPImageConverter instantiates a new ImageConverter
// which is able to parse HTTP requests and process
// images accordingly
func NewHTTPImageConverter() *HTTPImageConverter {
	return &HTTPImageConverter{
		Converter:    improc.NewImageConverter(),
		ParemeterMap: DefaultParameterMap(),
	}
}

// Read is the handler function for a HTTP request, and
// returns the raw image blob after the image has been
// processed
func (hic *HTTPImageConverter) Read(r *http.Request) ([]byte, error) {
	hic.Converter.Tracer(fmt.Sprintf("go-image-processor-http: parsing URL %s", r.URL.String()))

	pmap := hic.ParemeterMap
	if pmap == nil {
		hic.Converter.Tracer("go-image-processor-http: no parameter map available, using default")
		pmap = DefaultParameterMap()
	}

	preq, err := ParseURL(r.URL, pmap)
	if err != nil {
		hic.Converter.Tracer("go-image-processor-http: parsing URL failed!")
		return nil, err
	}

	hic.Converter.Tracer("go-image-processor-http: reading requested URL")

	reader := NewURLReader(preq.Source, pmap)
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
