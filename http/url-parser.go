package httpimproc

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	improc "github.com/ourstudio-se/go-image-processor"
)

// ProcessingRequest contains all necessary information to
// process an image for resizing, cropping, etc
type ProcessingRequest struct {
	Source     *url.URL
	OutputSpec *improc.OutputSpec
}

// ParseURL translates a HTTP URL with querystring, to a `ProcessingRequest`
func ParseURL(u *url.URL) (*ProcessingRequest, error) {
	query := u.Query()
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

	return nil, fmt.Errorf("only HTTP/S source URLs are supported")
}

func getFormatSpec(values url.Values) (*improc.OutputSpec, error) {
	spec := getParam(values, "spec")
	if spec != "" {
		return improc.ParseOutputSpec(spec)
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
		return improc.ParseOutputSpec(template)
	}
	if anchorX == "" || anchorY == "" {
		return nil, fmt.Errorf("malformed anchor specification")
	}

	template = fmt.Sprintf("%s@%s,%s", template, anchorX, anchorY)
	return improc.ParseOutputSpec(template)
}

func getCompression(values url.Values) improc.Compression {
	if outFormat := getParam(values, "out"); outFormat != "" {
		switch strings.ToLower(outFormat) {
		case "jpg", "jpeg":
			return improc.Jpeg
		case "png":
			return improc.Png
		case "webp":
			return improc.WebP
		}
	}

	return improc.TransitiveCompression
}

func getBackgroundColor(values url.Values, compression improc.Compression) improc.Color {
	if color := getParam(values, "background"); color != "" {
		_, err := strconv.ParseUint(color, 16, 64)
		if err == nil {
			return improc.Color(fmt.Sprintf("#%s", color))
		}
	}

	if compression == improc.Jpeg {
		return improc.Color("#FFFFFF")
	}

	return improc.ColorTransparent
}

func getParam(values url.Values, name string) string {
	v := values[name]
	if len(v) == 0 {
		return ""
	}

	return v[0]
}
