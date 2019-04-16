package improc

import (
	"math"

	"github.com/ourstudio-se/go-image-processor/abstractions"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type imageConverter struct {
	wand *imagick.MagickWand
}

// NewImageConverter creates a new converter
// which uses Imagick C bindings library
func NewImageConverter() abstractions.Converter {
	imagick.Initialize()

	return &imageConverter{}
}

func (c *imageConverter) Apply(blob []byte, spec *abstractions.OutputSpec) ([]byte, error) {
	c.wand = imagick.NewMagickWand()
	defer c.wand.Destroy()

	var err error

	err = c.fromBlob(blob)
	if err != nil {
		return nil, err
	}

	c.strip()
	err = c.applyFormat(spec)
	if err != nil {
		return nil, err
	}

	c.applyBackground(spec.Background, spec.Compression)

	bytes := c.bytes(spec.Quality, spec.Compression)
	return bytes, nil
}

func (c *imageConverter) fromBlob(blob []byte) error {
	err := c.wand.ReadImageBlob(blob)
	if err != nil {
		return err
	}

	c.wand.SetIteratorIndex(0)

	return nil
}

func (c *imageConverter) applyFormat(spec *abstractions.OutputSpec) error {
	var err error

	inputWidth := float64(c.wand.GetImageWidth())
	inputHeight := float64(c.wand.GetImageHeight())

	keepsRatio := (spec.Width / spec.Height) == (inputWidth / inputHeight)

	if spec.Width > 0 && spec.Height > 0 && !keepsRatio {
		if spec.Crop {
			if err = c.applyFormatWithCrop(inputWidth, inputHeight, spec); err != nil {
				return err
			}
		} else {
			if err = c.applyFormatWithoutCrop(inputWidth, inputHeight, spec); err != nil {
				return err
			}
		}
	} else {
		outputWidth := spec.Width
		outputHeight := spec.Height

		if spec.Width == 0 {
			outputWidth = math.Ceil((spec.Height / inputHeight) * inputWidth)
		}
		if spec.Height == 0 {
			outputHeight = math.Ceil((spec.Width / inputWidth) * inputHeight)
		}

		if err = c.wand.ResizeImage(uint(outputWidth), uint(outputHeight), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}
	}

	return nil
}

func (c *imageConverter) applyFormatWithoutCrop(inputWidth, inputHeight float64, spec *abstractions.OutputSpec) error {
	var err error

	isWiderThanHigher := (spec.Width / inputWidth) >= (spec.Height / inputHeight)
	isHigherThanWider := (spec.Width / inputWidth) < (spec.Height / inputHeight)

	if isWiderThanHigher {
		nextWidth := inputWidth * (spec.Height / inputHeight)
		if err = c.wand.ResizeImage(uint(nextWidth), uint(spec.Height), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}

		anchor := GetHorizontalAnchorValue(spec.Anchor, spec.Width, nextWidth)
		if err = c.wand.ExtentImage(uint(spec.Width), uint(spec.Height), anchor, 0); err != nil {
			return err
		}
	} else if isHigherThanWider {
		nextHeight := inputHeight * (spec.Width / inputWidth)
		if err = c.wand.ResizeImage(uint(spec.Width), uint(nextHeight), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}

		anchor := GetVerticalAnchorValue(spec.Anchor, spec.Height, nextHeight)
		if err = c.wand.ExtentImage(uint(spec.Width), uint(spec.Height), 0, anchor); err != nil {
			return err
		}
	}

	return nil
}

func (c *imageConverter) applyFormatWithCrop(inputWidth, inputHeight float64, spec *abstractions.OutputSpec) error {
	var err error

	isWiderThanHigher := (spec.Width / inputWidth) >= (spec.Height / inputHeight)
	isHigherThanWider := (spec.Width / inputWidth) < (spec.Height / inputHeight)

	if isHigherThanWider {
		nextWidth := inputWidth * (spec.Height / inputHeight)
		if err = c.wand.ResizeImage(uint(nextWidth), uint(spec.Height), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}

		anchor := GetHorizontalAnchorValue(spec.Anchor, spec.Width, nextWidth)
		if err = c.wand.CropImage(uint(spec.Width), uint(spec.Height), anchor, 0); err != nil {
			return err
		}
	} else if isWiderThanHigher {
		nextHeight := inputHeight * (spec.Width / inputWidth)
		if err = c.wand.ResizeImage(uint(spec.Width), uint(nextHeight), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}

		anchor := GetVerticalAnchorValue(spec.Anchor, spec.Height, nextHeight)
		if err = c.wand.CropImage(uint(spec.Width), uint(spec.Height), 0, anchor); err != nil {
			return err
		}
	}

	return nil
}

func (c *imageConverter) applyBackground(color abstractions.Color, compression abstractions.Compression) {
	if compression == abstractions.Jpeg {
		c.wand.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_REMOVE)

		if color == abstractions.ColorTransparent {
			color = abstractions.Color("#FFFFFF")
		}
	}
	if compression != abstractions.Jpeg && compression != abstractions.TransientCompression {
		c.wand.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_TRANSPARENT)
	}

	bg := imagick.NewPixelWand()
	defer bg.Destroy()

	bg.SetColor(color.String())
	c.wand.SetImageBackgroundColor(bg)
}

// GetHorizontalAnchorValue calculates where the anchor point should be relative to the
// current width (c) and the next/specified width (n) of the image
func GetHorizontalAnchorValue(a *abstractions.Anchor, c, n float64) int {
	if a.Horizontal == abstractions.GravityPull {
		// We "pull" the anchor to the left of the canvas
		return 0
	}
	if a.Horizontal == abstractions.GravityPush {
		// We "push" the anchor to the right of the canvas
		return int(-(c - n))
	}

	// Default gravity anchor point is the middle of the canvas
	return int(-(c - n) / 2)
}

// GetVerticalAnchorValue calculates where the anchor point should be relative to the
// current height (c) and the next/specified height (n) of the image
func GetVerticalAnchorValue(a *abstractions.Anchor, c, n float64) int {
	if a.Vertical == abstractions.GravityPull {
		// We "pull" the anchor to the top of the canvas
		return 0
	}
	if a.Vertical == abstractions.GravityPush {
		// We "push" the anchor to the bottom of the canvas
		return int(-(c - n))
	}

	// Default gravity anchor point is the middle of the canvas
	return int(-(c - n) / 2)
}

func (c *imageConverter) bytes(quality uint, compression abstractions.Compression) []byte {
	c.wand.SetImageCompressionQuality(quality)

	if compression != abstractions.TransientCompression && compression.String() != "" {
		c.wand.SetFormat(compression.String())
	}

	c.wand.ResetIterator()
	return c.wand.GetImageBlob()
}

func (c *imageConverter) strip() {
	c.wand.StripImage()
}

func (c *imageConverter) Destroy() {
	imagick.Terminate()
}
