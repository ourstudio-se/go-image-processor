package improc

import (
	"math"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type imagickWand interface {
	GetImageBlob() []byte
	ReadImageBlob([]byte) error
	GetImageWidth() uint
	GetImageHeight() uint
	CropImage(uint, uint, int, int) error
	ExtentImage(uint, uint, int, int) error
	ResizeImage(uint, uint, imagick.FilterType) error
	ResetIterator()
	SetFormat(string) error
	SetImageAlphaChannel(imagick.AlphaChannelType) error
	SetImageBackgroundColor(*imagick.PixelWand) error
	SetImageCompressionQuality(uint) error
	SetIteratorIndex(int) bool
	StripImage() error
	Destroy()
}

// ImageConverter handles output specifications and
// processes images to match the desired specification
type ImageConverter struct {
	wand imagickWand
}

// NewImageConverter creates a new converter
// which uses Imagick C bindings library
func NewImageConverter() *ImageConverter {
	imagick.Initialize()

	return &ImageConverter{}
}

// Apply takes an aoutput specification and processes
// the incoming image blob accordingly
func (c *ImageConverter) Apply(blob []byte, spec *OutputSpec) ([]byte, error) {
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

func (c *ImageConverter) fromBlob(blob []byte) error {
	err := c.wand.ReadImageBlob(blob)
	if err != nil {
		return err
	}

	c.wand.SetIteratorIndex(0)

	return nil
}

func (c *ImageConverter) applyFormat(spec *OutputSpec) error {
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

func (c *ImageConverter) applyFormatWithoutCrop(inputWidth, inputHeight float64, spec *OutputSpec) error {
	var err error

	isWiderThanHigher := (spec.Width / inputWidth) >= (spec.Height / inputHeight)
	isHigherThanWider := (spec.Width / inputWidth) < (spec.Height / inputHeight)

	if isWiderThanHigher {
		nextWidth := inputWidth * (spec.Height / inputHeight)
		if err = c.wand.ResizeImage(uint(nextWidth), uint(spec.Height), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}

		anchor := spec.Anchor.GetHorizontalAnchorValue(spec.Width, nextWidth)
		if err = c.wand.ExtentImage(uint(spec.Width), uint(spec.Height), anchor, 0); err != nil {
			return err
		}
	} else if isHigherThanWider {
		nextHeight := inputHeight * (spec.Width / inputWidth)
		if err = c.wand.ResizeImage(uint(spec.Width), uint(nextHeight), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}

		anchor := spec.Anchor.GetVerticalAnchorValue(spec.Height, nextHeight)
		if err = c.wand.ExtentImage(uint(spec.Width), uint(spec.Height), 0, anchor); err != nil {
			return err
		}
	}

	return nil
}

func (c *ImageConverter) applyFormatWithCrop(inputWidth, inputHeight float64, spec *OutputSpec) error {
	var err error

	isWiderThanHigher := (spec.Width / inputWidth) >= (spec.Height / inputHeight)
	isHigherThanWider := (spec.Width / inputWidth) < (spec.Height / inputHeight)

	if isHigherThanWider {
		nextWidth := inputWidth * (spec.Height / inputHeight)
		if err = c.wand.ResizeImage(uint(nextWidth), uint(spec.Height), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}

		anchor := spec.Anchor.GetHorizontalAnchorValue(spec.Width, nextWidth)
		if err = c.wand.CropImage(uint(spec.Width), uint(spec.Height), anchor, 0); err != nil {
			return err
		}
	} else if isWiderThanHigher {
		nextHeight := inputHeight * (spec.Width / inputWidth)
		if err = c.wand.ResizeImage(uint(spec.Width), uint(nextHeight), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}

		anchor := spec.Anchor.GetVerticalAnchorValue(spec.Height, nextHeight)
		if err = c.wand.CropImage(uint(spec.Width), uint(spec.Height), 0, anchor); err != nil {
			return err
		}
	}

	return nil
}

func (c *ImageConverter) applyBackground(color Color, compression Compression) {
	if compression == Jpeg {
		c.wand.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_REMOVE)

		if color == ColorTransparent {
			color = Color("#FFFFFF")
		}
	}
	if compression != Jpeg && compression != TransitiveCompression {
		c.wand.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_TRANSPARENT)
	}

	bg := imagick.NewPixelWand()
	defer bg.Destroy()

	bg.SetColor(color.String())
	c.wand.SetImageBackgroundColor(bg)
}

func (c *ImageConverter) bytes(quality uint, compression Compression) []byte {
	c.wand.SetImageCompressionQuality(quality)

	if compression != TransitiveCompression && compression.String() != "" {
		c.wand.SetFormat(compression.String())
	}

	c.wand.ResetIterator()
	return c.wand.GetImageBlob()
}

func (c *ImageConverter) strip() {
	c.wand.StripImage()
}

// Destroy terminates the ImageMagick session
func (c *ImageConverter) Destroy() {
	imagick.Terminate()
}
