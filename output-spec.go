package improc

import (
	"fmt"
	"strconv"
	"strings"
)

// Gravity defines an Enum for where a crop
// should anchor
type Gravity int

const (
	// GravityCenter enum value
	GravityCenter Gravity = 0

	// GravityPull enum value
	GravityPull Gravity = 1

	// GravityPush enum value
	GravityPush Gravity = 2
)

// Compression is an Enum specifying possible
// image compressions to use
type Compression int

const (
	// Jpeg image compression
	Jpeg Compression = iota

	// Png image compression
	Png

	// WebP image compression
	WebP

	// TransitiveCompression is using the
	// same compression algorithm as input source
	TransitiveCompression
)

func (c Compression) String() string {
	return [...]string{"jpg", "png", "webp", ""}[c]
}

// Color is a type definition for either a "none" value
// or for a hex numbered string
type Color string

// ColorTransparent defines a transparent color
const ColorTransparent Color = "none"

func (c Color) String() string {
	return (string)(c)
}

// Anchor defines where a resize or
// a crop should anchor, both horizontally
// and vertically
type Anchor struct {
	Horizontal Gravity
	Vertical   Gravity
}

// GetHorizontalAnchorValue calculates where the anchor point should be relative to the
// current width (c) and the next/specified width (n) of the image
func (a *Anchor) GetHorizontalAnchorValue(c, n float64) int {
	return getAnchorValue(a.Horizontal, c, n)
}

// GetVerticalAnchorValue calculates where the anchor point should be relative to the
// current height (c) and the next/specified height (n) of the image
func (a *Anchor) GetVerticalAnchorValue(c, n float64) int {
	return getAnchorValue(a.Vertical, c, n)
}

func getAnchorValue(g Gravity, c, n float64) int {
	if g == GravityPull {
		// We "pull" the anchor to the top/left of the canvas
		return 0
	}
	if g == GravityPush {
		// We "push" the anchor to the bottom/right of the canvas
		return int(-(c - n))
	}

	// Default gravity anchor point is the middle/center of the canvas
	return int(-(c - n) / 2)
}

// TextBlock defines a block of text to be applied
// to an image
type TextBlock struct {
	Text       string
	FontName   string
	FontSize   float64
	Foreground Color
	Background Color
	Anchor     *Anchor
}

type OverlaySource []byte

// OutputSpec is the specification used
// when resizing and/or cropping an image
type OutputSpec struct {
	Height      float64
	Width       float64
	Crop        bool
	Anchor      *Anchor
	Background  Color
	Overlays    []OverlaySource
	Quality     uint
	Compression Compression
	Text        *TextBlock
}

// ParseOutputSpec takes a string and returns a valid
// OutputSpec. The input string should include width and/or height
// plus an optional anchoring.
//
// Example: The string "200x100@1,1" represents an OutputSpec
// which should be 200px wide and 100px high. An anchor is also created
// for a resizing box rule, where the source image should be placed
// to the upper left corner when resizing the canvas.
func ParseOutputSpec(raw string) (*OutputSpec, error) {
	parts := strings.Split(strings.ToLower(raw), "@")
	anchor := &Anchor{
		Horizontal: GravityCenter,
		Vertical:   GravityCenter,
	}

	if len(parts) > 1 {
		anchor = ParseAnchorSpec(parts[1])
	}

	parts = strings.Split(parts[0], "x")
	if len(parts) != 2 {
		return nil, fmt.Errorf("the specified dimension format %s is not valid", raw)
	}

	w := 0
	h := 0
	var err error
	if parts[0] != "" && parts[1] != "" {
		w, err = strconv.Atoi(parts[0])
		h, err = strconv.Atoi(parts[1])
	} else if parts[0] != "" {
		w, err = strconv.Atoi(parts[0])
	} else if parts[1] != "" {
		h, err = strconv.Atoi(parts[1])
	}

	if err != nil {
		return nil, err
	}

	return &OutputSpec{
		Width:   float64(w),
		Height:  float64(h),
		Anchor:  anchor,
		Crop:    false,
		Quality: 85,
	}, nil
}

// ParseAnchorSpec returns horizontal and vertical anchoring from
// a string template, where values are separated with a comma.
func ParseAnchorSpec(raw string) *Anchor {
	parts := strings.Split(raw, ",")
	if len(parts) != 2 {
		return &Anchor{
			Horizontal: GravityCenter,
			Vertical:   GravityCenter,
		}
	}

	hz := GravityCenter
	vt := GravityCenter
	if x, err := strconv.Atoi(parts[0]); err == nil {
		if x < 0 {
			hz = GravityPull
		}
		if x > 0 {
			hz = GravityPush
		}
	}
	if x, err := strconv.Atoi(parts[1]); err == nil {
		if x < 0 {
			vt = GravityPull
		}
		if x > 0 {
			vt = GravityPush
		}
	}

	return &Anchor{
		Horizontal: hz,
		Vertical:   vt,
	}
}
