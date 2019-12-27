# go-image-processor

An easy to use image processor, with an optional HTTP handler to accept image conversion through a URL endpoint. It uses an underlying installation of imagemagick and Go bindings from [gographics](https://github.com/gographics/imagick).

## Why?

We often end up in projects where we have raw, or very high resolution, images that are served dynamically, that we need to resize or convert to a different format. Most solutions we've found have not met our needs fully - so wrapping imagemagick via Go, and applying our needed functionality, was the most viable option.

## Usage

### Dependencies

`go-image-processor` requires an existing installation of imagemagick v7.0+ and it's dev tools.

### Example

`go-image-processor` is only a library, but is very easy to implement - and contains all bits and bolts to create a runnable application in just a few lines of code.

The following example takes an image file and resizes it to an output specification, specified as a string template.

```go
package main

import (
	"github.com/ourstudio-se/go-image-processor"
)

func main() {
	converter := improc.NewImageConverter()
	defer converter.Destroy()

	b, err := ioutil.ReadFile("image.png")
	if err != nil {
		panic(err)
	}

	spec, err := improc.ParseOutputSpec("200x200")
	if err != nil {
		panic(err)
	}

	output, err := converter.Apply(b, spec)
	if err != nil {
		panic(err)
	}

	err := ioutil.WriteFile("image-200x200.png", output, 0644)
	if err != nil {
		panic(err)
	}
}
```

## HTTP request handler

The library includes HTTP functionality, which takes a `*http.Request` and reads querystring values to determine what actions to take. It can resize, crop, and rewrite images on the fly.

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	httpimproc "github.com/ourstudio-se/go-image-processor/http"
)

type httpapi struct {
	conv *httpimproc.HTTPImageConverter
}

func (ha *httpapi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := ha.conv.Read(r)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(b)))
	w.Header().Set("Content-Disposition", "inline")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func main() {
	conv := httpimproc.NewHTTPImageConverter()
	api := &httpapi{
		conv,
	}

	http.Handle("/convert", api)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

The querystring parameters available are defined below.

### `url`

The source image URL.

### `spec`

A string template for shortcuts to desired output specifications, reducing the number of required query parameters. The value can take different shapes:
- `100x200`: Shortcut for `width=100&height=200`
- `100x`: Shortcut for `width=100` and dynamic height (to scale)
- `x200`: Shortcut for `height=200` and dynamic width (to scale)
- `100x200@-1,1`: Shortcut for `width=100&height=200&anchorx=-1&anchory=1`

### `out`

Specifies which compression the output image should have, it can be `jpg`, `png`, or `webp`. By default it's transitive, meaning the input compression defines the output compression - if the source image is a JPEG image and no `out` parameter is specified, the output compression would be JPEG as well.

### `width`

Specifies the desired output width of an image.

### `height`

Specifies the desired output height of an image.

### `anchorx`

When requesting a resize without cropping an image (e.g. no image data loss), one can specify where the output data should be placed on the canvas horizontally by specifying `anchorx`. It can take three values:
- `0` (default) centerizes
- `-1` pulls the image to the left edge of the canvas
- `1` pushes the image to the right edge of the canvas

### `anchory`

As `anchorx`, but defines the vertical alignment on the canvas:
- `0` (default) centerizes
- `-1` pulls the image to the top edge of the canvas
- `1` pushes the image to the bottom edge of the canvas

### `crop`

Specifies if the image should be cropped (with possible image data loss) or not. Valid values are `true` and `false`.

### `quality`

Specifies output quality, with a value between 0 and 100. Defaults to 85.

### `background`

Apply a background color for images where the canvas is visible (e.g. after a non cropped resize). Input values should be in hex format, such as `FF00BB`. Defaults to white for JPEG outputs and defaults to transparent for PNG/WebP.

### `text:value`

A text block to be applied to the output image. Only applicable when `text:font` and `text:size` are set as well.

### `text:font`

A font for a text block to be applied to the output image. Only applicable when `text:value` and `text:size` are set as well.

### `text:size`

A font size for a text block to be applied to the output image. Only applicable when `text:value` and `text:font` are set as well.

### `text:foreground`

Specifies a font color for a text block. Values should be in hex format, such as `000000`. Defaults to black.

### `text:background`

Specifies a background color for a text block. Values should be in hex format, such as `FFFFFF`. Defaults to transparent.

### `text:anchor`

Specifies an anchor point for a text block. Valid values are:

- `0,0`: Center/Center
- `-1,-1`: Upper/Left
- `-1,1`: Upper/Right
- `1,-1`: Lower/Left
- `1,1`: Lower/Right
- `0,-1`: Center/Left
- `0,1`: Center/Right
- `-1,0`: Upper/Center
- `1,0`: Lower/Center

## License

MIT