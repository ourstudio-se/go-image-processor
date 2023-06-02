package improc

import (
	"gopkg.in/gographics/imagick.v3/imagick"
)

// ImageConverter handles output specifications and
// processes images to match the desired specification
type ImageConverter struct{}

// NewImageConverter creates a new converter
// which uses Imagick C bindings library
func NewImageConverter() *ImageConverter {
	imagick.Initialize()

	return &ImageConverter{}
}

// Apply takes an aoutput specification and processes
// the incoming image blob accordingly
func (c *ImageConverter) Apply(blob []byte, spec *OutputSpec) ([]byte, error) {
	h := newHandler()
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
	imagick.Terminate()
}
