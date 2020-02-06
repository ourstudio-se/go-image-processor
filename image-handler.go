package improc

import (
	"fmt"
	"math"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type handler struct {
	pool pool
	wand *imagick.MagickWand
}

func newHandler(p pool) (*handler, error) {
	w, err := p.Take()
	if err != nil {
		return nil, fmt.Errorf("could not acquire wand")
	}

	return &handler{
		pool: p,
		wand: w,
	}, nil
}

func (h *handler) fromBlob(blob []byte) error {
	err := h.wand.ReadImageBlob(blob)
	if err != nil {
		return err
	}

	h.wand.SetIteratorIndex(0)

	return nil
}

func (h *handler) applyFormat(spec *OutputSpec) error {
	var err error

	inputWidth := float64(h.wand.GetImageWidth())
	inputHeight := float64(h.wand.GetImageHeight())
	keepsRatio := (spec.Width / spec.Height) == (inputWidth / inputHeight)

	if spec.Width > 0 && spec.Height > 0 && !keepsRatio {
		if spec.Crop {
			if err = h.applyFormatWithCrop(inputWidth, inputHeight, spec); err != nil {
				return err
			}
		} else {
			if err = h.applyFormatWithoutCrop(inputWidth, inputHeight, spec); err != nil {
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

		if err = h.wand.ResizeImage(uint(outputWidth), uint(outputHeight), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) applyFormatWithoutCrop(inputWidth, inputHeight float64, spec *OutputSpec) error {
	var err error

	isWiderThanHigher := (spec.Width / inputWidth) >= (spec.Height / inputHeight)
	isHigherThanWider := (spec.Width / inputWidth) < (spec.Height / inputHeight)

	if isWiderThanHigher {
		nextWidth := inputWidth * (spec.Height / inputHeight)

		if err = h.wand.ResizeImage(uint(nextWidth), uint(spec.Height), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}

		anchor := spec.Anchor.GetHorizontalAnchorValue(spec.Width, nextWidth)

		if err = h.wand.ExtentImage(uint(spec.Width), uint(spec.Height), anchor, 0); err != nil {
			return err
		}
	} else if isHigherThanWider {
		nextHeight := inputHeight * (spec.Width / inputWidth)

		if err = h.wand.ResizeImage(uint(spec.Width), uint(nextHeight), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}

		anchor := spec.Anchor.GetVerticalAnchorValue(spec.Height, nextHeight)

		if err = h.wand.ExtentImage(uint(spec.Width), uint(spec.Height), 0, anchor); err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) applyFormatWithCrop(inputWidth, inputHeight float64, spec *OutputSpec) error {
	var err error

	isWiderThanHigher := (spec.Width / inputWidth) >= (spec.Height / inputHeight)
	isHigherThanWider := (spec.Width / inputWidth) < (spec.Height / inputHeight)

	if isHigherThanWider {
		nextWidth := inputWidth * (spec.Height / inputHeight)

		if err = h.wand.ResizeImage(uint(nextWidth), uint(spec.Height), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}

		anchor := spec.Anchor.GetHorizontalAnchorValue(spec.Width, nextWidth)

		if err = h.wand.CropImage(uint(spec.Width), uint(spec.Height), anchor, 0); err != nil {
			return err
		}
	} else if isWiderThanHigher {
		nextHeight := inputHeight * (spec.Width / inputWidth)

		if err = h.wand.ResizeImage(uint(spec.Width), uint(nextHeight), imagick.FILTER_LANCZOS2); err != nil {
			return err
		}

		anchor := spec.Anchor.GetVerticalAnchorValue(spec.Height, nextHeight)

		if err = h.wand.CropImage(uint(spec.Width), uint(spec.Height), 0, anchor); err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) applyBackground(color Color, compression Compression) {
	if compression == Jpeg {
		h.wand.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_REMOVE)

		if color == ColorTransparent {
			color = Color("#FFFFFF")
		}
	}
	if compression != Jpeg && compression != TransitiveCompression {
		h.wand.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_SET)
	}

	bg := imagick.NewPixelWand()
	defer bg.Destroy()

	bg.SetColor(color.String())
	h.wand.SetImageBackgroundColor(bg)
}

func (h *handler) applyTextBlock(tb *TextBlock) error {
	var err error

	mw, err := h.pool.Take()
	if err != nil {
		return err
	}
	defer h.pool.Put(mw)

	dw := imagick.NewDrawingWand()
	fg := imagick.NewPixelWand()
	bg := imagick.NewPixelWand()

	fg.SetColor(tb.Foreground.String())
	bg.SetColor(tb.Background.String())

	dw.SetFillColor(fg)
	dw.SetFontSize(tb.FontSize)

	if err = dw.SetFont(tb.FontName); err != nil {
		return err
	}

	if err = mw.NewImage(h.wand.GetImageWidth(), h.wand.GetImageHeight(), bg); err != nil {
		return err
	}

	fm := mw.QueryFontMetrics(dw, "W")
	dy := fm.CharacterHeight + fm.Descender

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

	nextWidth := int(mw.GetImageWidth()) + int(tb.FontSize)
	nextHeight := int(mw.GetImageHeight()) + int(dy)

	anchorX := -(nextWidth - int(mw.GetImageWidth())) / 2
	anchorY := -(nextHeight - int(mw.GetImageHeight()) - int(fm.Descender)) / 2

	if err = mw.ExtentImage(uint(nextWidth), uint(nextHeight), anchorX, anchorY); err != nil {
		return err
	}

	x, y := 0, 0
	if tb.Anchor.Horizontal == GravityPull {
		x = 0
	}
	if tb.Anchor.Horizontal == GravityCenter {
		x = int((h.wand.GetImageWidth() / 2) - (mw.GetImageWidth() / 2))
	}
	if tb.Anchor.Horizontal == GravityPush {
		x = int(h.wand.GetImageWidth() - mw.GetImageWidth())
	}
	if tb.Anchor.Vertical == GravityPull {
		y = 0
	}
	if tb.Anchor.Vertical == GravityCenter {
		y = int((h.wand.GetImageHeight() / 2) - (mw.GetImageHeight() / 2))
	}
	if tb.Anchor.Vertical == GravityPush {
		y = int(h.wand.GetImageHeight() - mw.GetImageHeight())
	}

	return h.wand.CompositeImage(mw, imagick.COMPOSITE_OP_OVER, true, x, y)
}

func (h *handler) bytes(quality uint, compression Compression) []byte {
	h.wand.SetImageCompressionQuality(quality)

	if compression != TransitiveCompression && compression.String() != "" {
		h.wand.SetImageFormat(compression.String())
	}

	h.wand.ResetIterator()
	b := h.wand.GetImageBlob()
	return b
}

func (h *handler) strip() {
	h.wand.StripImage()
}

func (h *handler) destroy() {
	h.pool.Put(h.wand)
}
