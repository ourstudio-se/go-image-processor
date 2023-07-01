package improc

import (
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

// ImageConverter handles output specifications and
// processes images to match the desired specification
type ImageConverter struct {
	pool pool
}

// NewImageConverter creates a new converter
// which uses Imagick C bindings library
func NewImageConverter() (*ImageConverter, error) {
	imagick.Initialize()

	pool, err := newWandPool(100)
	if err != nil {
		return nil, err
	}

	return &ImageConverter{
		pool,
	}, nil
}

// Apply takes an aoutput specification and processes
// the incoming image blob accordingly
func (c *ImageConverter) Apply(blob []byte, spec *OutputSpec) ([]byte, error) {
	var err error

	h, err := newHandler(c.pool)
	if err != nil {
		return nil, fmt.Errorf("failed creating image handler: %v", err)
	}

	defer h.destroy()

	err = h.fromBlob(blob)
	if err != nil {
		return nil, err
	}

	h.strip()
	err = h.applyFormat(spec)
	if err != nil {
		return nil, err
	}

	h.applyBackground(spec.Background, spec.Compression)

	if len(spec.Overlays) > 0 {
		for _, overlay := range spec.Overlays {
			h.addLayer(overlay)
		}
	}

	if spec.Text != nil {
		if err = h.applyTextBlock(spec.Text); err != nil {
			return nil, err
		}
	}

	bytes := h.bytes(spec.Quality, spec.Compression)

	return bytes, nil
}

// Destroy terminates the ImageMagick session
func (c *ImageConverter) Destroy() {
	c.pool.Close()
	imagick.Terminate()
}
