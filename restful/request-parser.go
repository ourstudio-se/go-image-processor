package restful

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ourstudio-se/go-image-processor/abstractions"
)

// ProcessingRequest contains all necessary information to
// process an image for resizing, cropping, etc
type ProcessingRequest struct {
	Source     *url.URL
	OutputSpec *abstractions.OutputSpec
}

// NewProcessingRequest translates a GET request to a `processingRequest`
func NewProcessingRequest(r *http.Request) (*ProcessingRequest, error) {
	query := r.URL.Query()
	source, err := getImageSource(query)
	if err != nil {
		return nil, err
	}

	formatSpec, err := getFormatSpec(query)
	if err != nil {
		return nil, err
	}

	if getParam(query, "crop") == "true" {
		formatSpec.Crop = true
	}
	if q, err := strconv.Atoi(getParam(query, "quality")); err == nil && q > 0 && q <= 100 {
		formatSpec.Quality = uint(q)
	}
	formatSpec.Compression = getCompression(query)
	formatSpec.Background = getBackgroundColor(query, formatSpec.Compression)

	return &ProcessingRequest{
		Source:     source,
		OutputSpec: formatSpec,
	}, nil
}

func getImageSource(values url.Values) (*url.URL, error) {
	source, err := url.Parse(getParam(values, "url"))
	if err != nil {
		return nil, err
	}

	if !source.IsAbs() {
		return nil, fmt.Errorf("the source URL '%s' is malformed", source.String())
	}
	switch strings.ToLower(source.Scheme) {
	case "http", "https":
		return source, nil
	}

	return nil, fmt.Errorf("only HTTP and HTTPS source URLs are supported")
}

func getFormatSpec(values url.Values) (*abstractions.OutputSpec, error) {
	spec := getParam(values, "spec")
	if spec != "" {
		return abstractions.ParseOutputSpec(spec)
	}

	width := getParam(values, "width")
	height := getParam(values, "height")

	if width == "" && height == "" {
		return nil, fmt.Errorf("missing output dimensions")
	}

	template := fmt.Sprintf("%sx%s", width, height)

	anchorX := getParam(values, "anchorx")
	anchorY := getParam(values, "anchory")

	if anchorX == "" && anchorY == "" {
		return abstractions.ParseOutputSpec(template)
	}
	if anchorX == "" || anchorY == "" {
		return nil, fmt.Errorf("malformed anchor specifications")
	}

	template = fmt.Sprintf("%s@%s,%s", template, anchorX, anchorY)
	return abstractions.ParseOutputSpec(template)
}

func getCompression(values url.Values) abstractions.Compression {
	if outFormat := getParam(values, "out"); outFormat != "" {
		switch strings.ToLower(outFormat) {
		case "jpg", "jpeg":
			return abstractions.Jpeg
		case "png":
			return abstractions.Png
		case "webp":
			return abstractions.WebP
		}
	}

	return abstractions.TransientCompression
}

func getBackgroundColor(values url.Values, compression abstractions.Compression) abstractions.Color {
	if color := getParam(values, "background"); color != "" {
		_, err := strconv.ParseUint(color, 16, 64)
		if err != nil {
			return abstractions.Color(fmt.Sprintf("#%s", color))
		}
	}

	if compression == abstractions.Jpeg {
		return abstractions.Color("#FFFFFF")
	}

	return abstractions.ColorTransparent
}

func getParam(values url.Values, name string) string {
	v := values[name]
	if len(v) == 0 {
		return ""
	}

	return v[0]
}
