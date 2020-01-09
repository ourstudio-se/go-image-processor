package improc

import (
	"fmt"
	"math"

	"gopkg.in/gographics/imagick.v3/imagick"
)

// ImageConverter handles output specifications and
// processes images to match the desired specification
type ImageConverter struct {
	wand   *imagick.MagickWand
	Tracer func(s string)
}

func noOpTracer(_ string) {}

// NewImageConverter creates a new converter
// which uses Imagick C bindings library
func NewImageConverter() *ImageConverter {
	imagick.Initialize()

	return &ImageConverter{
		Tracer: noOpTracer,
	}
}

// Apply takes an aoutput specification and processes
// the incoming image blob accordingly
func (c *ImageConverter) Apply(blob []byte, spec *OutputSpec) ([]byte, error) {
	c.wand = imagick.NewMagickWand()
	c.Tracer("go-image-processor: new MagickWand created")
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

	if spec.Text != nil {
		if err = c.applyTextBlock(spec.Text); err != nil {
			c.Tracer("go-image-processor: could not apply text block!")
		}

	}

	bytes := c.bytes(spec.Quality, spec.Compression)

	c.Tracer(fmt.Sprintf("go-image-processor: MagickWand completed, returning %d bytes to caller", len(bytes)))
	return bytes, nil
}

func (c *ImageConverter) fromBlob(blob []byte) error {
	c.Tracer("go-image-processor: loading byte blob into MagickWand")
	err := c.wand.ReadImageBlob(blob)
	if err != nil {
		c.Tracer("go-image-processor: failed loading bytes into MagickWand")
		return err
	}

	c.Tracer("go-image-processor: resetting iterator index to 0")
	c.wand.SetIteratorIndex(0)

	return nil
}

func (c *ImageConverter) applyFormat(spec *OutputSpec) error {
	var err error

	inputWidth := float64(c.wand.GetImageWidth())
	c.Tracer(fmt.Sprintf("go-image-processor: loading input width from original image: %.0f", inputWidth))

	inputHeight := float64(c.wand.GetImageHeight())
	c.Tracer(fmt.Sprintf("go-image-processor: loading input height from originl image: %.0f", inputHeight))

	keepsRatio := (spec.Width / spec.Height) == (inputWidth / inputHeight)

	c.Tracer(fmt.Sprintf("go-image-processor: keeping ratio: %t", keepsRatio))

	if spec.Width > 0 && spec.Height > 0 && !keepsRatio {
		if spec.Crop {
			if err = c.applyFormatWithCrop(inputWidth, inputHeight, spec); err != nil {
				c.Tracer("go-image-processor: applying spec with crop failed!")
				return err
			}
		} else {
			if err = c.applyFormatWithoutCrop(inputWidth, inputHeight, spec); err != nil {
				c.Tracer("go-image-processor: applying spec without crop failed!")
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

		c.Tracer(fmt.Sprintf("go-image-processor: resizing image to %.0fx%.0f according to spec", outputWidth, outputHeight))

		if err = c.wand.ResizeImage(uint(outputWidth), uint(outputHeight), imagick.FILTER_LANCZOS2); err != nil {
			c.Tracer("go-image-processor: resizing image failed!")
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

		c.Tracer(fmt.Sprintf("go-image-processor: no crop: resizing image to %.0fx%.0f before applying extent", nextWidth, spec.Height))

		if err = c.wand.ResizeImage(uint(nextWidth), uint(spec.Height), imagick.FILTER_LANCZOS2); err != nil {
			c.Tracer("go-image-processor: no crop: resizing image failed!")
			return err
		}

		anchor := spec.Anchor.GetHorizontalAnchorValue(spec.Width, nextWidth)

		c.Tracer(fmt.Sprintf("go-image-processor: no crop: using horizontal anchor %d when creating extent %.0fx%.0f", anchor, spec.Width, spec.Height))

		if err = c.wand.ExtentImage(uint(spec.Width), uint(spec.Height), anchor, 0); err != nil {
			c.Tracer("go-image-processor: no crop: extending image canvas failed!")
			return err
		}
	} else if isHigherThanWider {
		nextHeight := inputHeight * (spec.Width / inputWidth)

		c.Tracer(fmt.Sprintf("go-image-processor: no crop: resizing image to %.0fx%.0f before applying extent", spec.Width, nextHeight))

		if err = c.wand.ResizeImage(uint(spec.Width), uint(nextHeight), imagick.FILTER_LANCZOS2); err != nil {
			c.Tracer("go-image-processor: no crop: resizing image failed!")
			return err
		}

		anchor := spec.Anchor.GetVerticalAnchorValue(spec.Height, nextHeight)

		c.Tracer(fmt.Sprintf("go-image-processor: no crop: using vertical anchor %d when creating extent %.0fx%.0f", anchor, spec.Width, spec.Height))

		if err = c.wand.ExtentImage(uint(spec.Width), uint(spec.Height), 0, anchor); err != nil {
			c.Tracer("go-image-processor: no crop: extending image canvas failed!")
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

		c.Tracer(fmt.Sprintf("go-image-processor: with crop: resizing image to %.0fx%.0f before applying extent", nextWidth, spec.Height))

		if err = c.wand.ResizeImage(uint(nextWidth), uint(spec.Height), imagick.FILTER_LANCZOS2); err != nil {
			c.Tracer("go-image-processor: with crop: resizing image failed!")
			return err
		}

		anchor := spec.Anchor.GetHorizontalAnchorValue(spec.Width, nextWidth)

		c.Tracer(fmt.Sprintf("go-image-processor: with crop: using horizontal anchor %d when cropping to %.0fx%.0f", anchor, spec.Width, spec.Height))

		if err = c.wand.CropImage(uint(spec.Width), uint(spec.Height), anchor, 0); err != nil {
			c.Tracer("go-image-processor: with crop: cropping image canvas failed!")
			return err
		}
	} else if isWiderThanHigher {
		nextHeight := inputHeight * (spec.Width / inputWidth)

		c.Tracer(fmt.Sprintf("go-image-processor: with crop: resizing image to %.0fx%.0f before applying extent", spec.Width, nextHeight))

		if err = c.wand.ResizeImage(uint(spec.Width), uint(nextHeight), imagick.FILTER_LANCZOS2); err != nil {
			c.Tracer("go-image-processor: with crop: resizing image failed!")
			return err
		}

		anchor := spec.Anchor.GetVerticalAnchorValue(spec.Height, nextHeight)

		c.Tracer(fmt.Sprintf("go-image-processor: with crop: using vertical anchor %d when cropping to %.0fx%.0f", anchor, spec.Width, spec.Height))

		if err = c.wand.CropImage(uint(spec.Width), uint(spec.Height), 0, anchor); err != nil {
			c.Tracer("go-image-processor: with crop: cropping image canvas failed!")
			return err
		}
	}

	return nil
}

func (c *ImageConverter) applyBackground(color Color, compression Compression) {
	if compression == Jpeg {
		c.Tracer("go-image-processor: removing alpha channel")

		c.wand.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_REMOVE)

		if color == ColorTransparent {
			color = Color("#FFFFFF")
		}
	}
	if compression != Jpeg && compression != TransitiveCompression {
		c.Tracer("go-image-processor: setting alpha channel")

		c.wand.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_SET)
	}

	bg := imagick.NewPixelWand()
	defer bg.Destroy()

	c.Tracer(fmt.Sprintf("go-image-processor: setting background color %s", color.String()))

	bg.SetColor(color.String())
	c.wand.SetImageBackgroundColor(bg)
}

func (c *ImageConverter) applyTextBlock(tb *TextBlock) error {
	c.Tracer("go-image-processor: adding text block")
	var err error

	mw := imagick.NewMagickWand()
	dw := imagick.NewDrawingWand()
	fg := imagick.NewPixelWand()
	bg := imagick.NewPixelWand()

	defer mw.Destroy()
	defer dw.Destroy()
	defer fg.Destroy()
	defer bg.Destroy()

	c.Tracer(fmt.Sprintf("go-image-processor: setting text block foreground color to %s", tb.Foreground.String()))
	fg.SetColor(tb.Foreground.String())

	c.Tracer(fmt.Sprintf("go-image-processor: setting text block background color to %s", tb.Background.String()))
	bg.SetColor(tb.Background.String())

	dw.SetFillColor(fg)

	c.Tracer(fmt.Sprintf("go-image-processor: setting text block font size to %.0f", tb.FontSize))
	dw.SetFontSize(tb.FontSize)

	c.Tracer(fmt.Sprintf("go-image-processor: setting text block font to %s", tb.FontName))
	if err = dw.SetFont(tb.FontName); err != nil {
		return err
	}

	if err = mw.NewImage(c.wand.GetImageWidth(), c.wand.GetImageHeight(), bg); err != nil {
		return err
	}

	fm := mw.QueryFontMetrics(dw, "W")
	dy := (fm.CharacterHeight * 1.5) + fm.Descender

	c.Tracer(fmt.Sprintf("go-image-processor: setting text block value to %s", tb.Text))
	dw.Annotation(10, dy, tb.Text)

	if err = mw.DrawImage(dw); err != nil {
		return err
	}
	if err = mw.TrimImage(0); err != nil {
		return err
	}
	if err = mw.SetImageBackgroundColor(bg); err != nil {
		return err
	}

	pad := int(tb.FontSize)
	nextWidth := int(mw.GetImageWidth()) + pad
	nextHeight := int(dy) - (int(dy) / 2) + pad

	c.Tracer(fmt.Sprintf("go-image-processor: setting text block size to %dx%d", nextWidth, nextHeight))
	if err = mw.ExtentImage(uint(nextWidth), uint(nextHeight), -(nextWidth-int(mw.GetImageWidth()))/2, -(int(dy))/2); err != nil {
		return err
	}

	x, y := 0, 0
	if tb.Anchor.Horizontal == GravityPull {
		x = 0
	}
	if tb.Anchor.Horizontal == GravityCenter {
		x = int((c.wand.GetImageWidth() / 2) - (mw.GetImageWidth() / 2))
	}
	if tb.Anchor.Horizontal == GravityPush {
		x = int(c.wand.GetImageWidth() - mw.GetImageWidth())
	}
	if tb.Anchor.Vertical == GravityPull {
		y = 0
	}
	if tb.Anchor.Vertical == GravityCenter {
		y = int((c.wand.GetImageHeight() / 2) - (mw.GetImageHeight() / 2))
	}
	if tb.Anchor.Vertical == GravityPush {
		y = int(c.wand.GetImageHeight() - mw.GetImageHeight())
	}

	return c.wand.CompositeImage(mw, imagick.COMPOSITE_OP_OVER, true, x, y)
}

func (c *ImageConverter) bytes(quality uint, compression Compression) []byte {
	c.Tracer(fmt.Sprintf("go-image-processor: setting compression quality to %d", quality))

	c.wand.SetImageCompressionQuality(quality)

	if compression != TransitiveCompression && compression.String() != "" {
		c.Tracer(fmt.Sprintf("go-image-processor: setting compression to %s", compression.String()))

		c.wand.SetImageFormat(compression.String())
	}

	c.Tracer("go-image-processor: resetting iterator")
	c.wand.ResetIterator()

	c.Tracer("go-image-processor: reading byte blob from MagickWand")
	b := c.wand.GetImageBlob()
	return b
}

func (c *ImageConverter) strip() {
	c.Tracer("go-image-processor: stripping exif data")
	c.wand.StripImage()
}

// Destroy terminates the ImageMagick session
func (c *ImageConverter) Destroy() {
	imagick.Terminate()
}
