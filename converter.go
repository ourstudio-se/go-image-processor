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
func NewImageConverter() *ImageConverter {
	imagick.Initialize()

	pool := newWandPool(100)
	return &ImageConverter{
		pool,
	}
}

// Apply takes an aoutput specification and processes
// the incoming image blob accordingly
func (c *ImageConverter) Apply(blob []byte, spec *OutputSpec) ([]byte, error) {
	h, err := newHandler(c.pool)
	if err != nil {
		return nil, fmt.Errorf("failed creating image handler: %v", err)
	}

	defer h.destroy()

	var err error

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
