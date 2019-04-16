# go-image-processor

An easy to use image processor, with an optional HTTP handler to accept image conversion through a URL endpoint. It uses an underlying installation of imagemagick and Go bindings from [gographics](https://github.com/gographics/imagick).

## Why?

We often end up in projects where we have raw, or very high resolution, images that are served dynamically, that we need to resize or convert to a different format. Most solutions we've found have not met our needs fully - so wrapping imagemagick via Go, and applying our needed functionality, was the most viable option.

## Usage

### Dependencies

`go-image-processor` requires an existing installation of imagemagick v7.0+ and it's dev tools.

### Example

`go-image-processor` is only a library, but is very easy to implement - and contains all bits and bolts to create a runnable application in just a few lines of code.

The following example starts a webserver on port 8080, exposing a single endpoint (http://localhost:8080/convert) for image conversion. In the [example directory](https://github.com/ourstudio-se/go-image-processor/tree/master/example) there's a fully runnable, Dockerized, application.

```go
package main

import (
	"github.com/ourstudio-se/go-image-processor"
	"github.com/ourstudio-se/go-image-processor/abstractions"
	"github.com/ourstudio-se/go-image-processor/readers"
	"github.com/ourstudio-se/go-image-processor/restful"

	"go.uber.org/dig"
)

func main() {
	container := dig.New()
	container.Provide(improc.NewImageConverter)

	container.Provide(readers.NewURLReaderOptions)
	container.Provide(readers.NewURLReaderFactory)

	container.Provide(func () *restful.APIOptions {
		return restful.NewAPIOptions("/convert")
	})
	container.Provide(restful.NewAPI)

	// API.Start() is blocking
	container.Invoke(func(api *restful.API) error {
		return api.Start()
	})

	// Clean up if the API server is closed
	container.Invoke(func(converter abstractions.Converter) {
		converter.Destroy()
	})
}
```

## Restful query API

The restful API requires at least one dimenson (width or height) and a source URL to convert an image.

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

## License

MIT