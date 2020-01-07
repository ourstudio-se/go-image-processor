package httpimproc

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	improc "github.com/ourstudio-se/go-image-processor/v2"
)

// ProcessingRequest contains all necessary information to
// process an image for resizing, cropping, etc
type ProcessingRequest struct {
	Source     *url.URL
	OutputSpec *improc.OutputSpec
}

// ParseURL translates a HTTP URL with querystring, to a `ProcessingRequest`
func ParseURL(u *url.URL, parameters *ParameterMap) (*ProcessingRequest, error) {
	query := u.Query()
	source, err := getImageSource(query, parameters.SourceURL)
	if err != nil {
		return nil, err
	}

	formatSpec, err := getFormatSpec(query, parameters)
	if err != nil {
		return nil, err
	}

	if getParam(query, parameters.Crop) == "true" {
		formatSpec.Crop = true
	}
	if q, err := strconv.Atoi(getParam(query, parameters.Quality)); err == nil && q > 0 && q <= 100 {
		formatSpec.Quality = uint(q)
	}
	formatSpec.Compression = getCompression(query, parameters.Compression)
	formatSpec.Background = getBackgroundColor(query, formatSpec.Compression, parameters.Background)
	formatSpec.Text = getTextBlock(query, parameters)

	return &ProcessingRequest{
		Source:     source,
		OutputSpec: formatSpec,
	}, nil
}

func getImageSource(values url.Values, param string) (*url.URL, error) {
	source, err := url.Parse(getParam(values, param))
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

func getFormatSpec(values url.Values, parameters *ParameterMap) (*improc.OutputSpec, error) {
	spec := getParam(values, parameters.Spec)
	if spec != "" {
		return improc.ParseOutputSpec(spec)
	}

	width := getParam(values, parameters.Width)
	height := getParam(values, parameters.Height)

	if width == "" && height == "" {
		return nil, fmt.Errorf("missing output dimensions")
	}

	template := fmt.Sprintf("%sx%s", width, height)

	anchorX := getParam(values, parameters.AnchorX)
	anchorY := getParam(values, parameters.AnchorY)

	if anchorX == "" && anchorY == "" {
		return improc.ParseOutputSpec(template)
	}
	if anchorX == "" || anchorY == "" {
		return nil, fmt.Errorf("malformed anchor specification")
	}

	template = fmt.Sprintf("%s@%s,%s", template, anchorX, anchorY)
	return improc.ParseOutputSpec(template)
}

func getCompression(values url.Values, param string) improc.Compression {
	if outFormat := getParam(values, param); outFormat != "" {
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

func getTextBlock(values url.Values, parameters *ParameterMap) *improc.TextBlock {
	tb := &improc.TextBlock{
		Foreground: improc.Color("#000000"),
		Background: improc.Color("none"),
		Anchor: &improc.Anchor{
			Horizontal: improc.GravityCenter,
			Vertical:   improc.GravityCenter,
		},
	}

	if text := getParam(values, parameters.TextValue); text != "" {
		tb.Text = text
	} else {
		return nil
	}

	if font := getParam(values, parameters.TextFont); font != "" {
		tb.FontName = font
	} else {
		return nil
	}

	if size := getParam(values, parameters.TextSize); size != "" {
		fontSize, err := strconv.ParseFloat(size, 64)
		if err == nil && fontSize > 0 {
			tb.FontSize = fontSize
		} else {
			return nil
		}
	} else {
		return nil
	}

	if fg := getParam(values, parameters.TextForeground); fg != "" {
		fgColor, err := getColor(fg)
		if err == nil {
			tb.Foreground = fgColor
		}
	}

	if bg := getParam(values, parameters.TextBackground); bg != "" {
		bgColor, err := getColor(bg)
		if err == nil {
			tb.Background = bgColor
		}
	}

	if anchors := getParam(values, parameters.TextAnchor); anchors != "" {
		tb.Anchor = improc.ParseAnchorSpec(anchors)
	}

	return tb
}

func getBackgroundColor(values url.Values, compression improc.Compression, param string) improc.Color {
	if color := getParam(values, param); color != "" {
		c, err := getColor(color)
		if err == nil {
			return c
		}
	}

	if compression == improc.Jpeg {
		return improc.Color("#FFFFFF")
	}

	return improc.ColorTransparent
}

func getColor(c string) (improc.Color, error) {
	_, err := strconv.ParseUint(c, 16, 64)
	if err != nil {
		return improc.ColorTransparent, err
	}

	return improc.Color(fmt.Sprintf("#%s", c)), nil
}

func getParam(values url.Values, name string) string {
	v := values[name]
	if len(v) == 0 {
		return ""
	}

	return v[0]
}
