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

	container.Provide(func() *restful.APIOptions {
		return restful.NewAPIOptions("/convert")
	})
	container.Provide(restful.NewAPI)

	container.Invoke(func(api *restful.API) error {
		return api.Start()
	})

	container.Invoke(func(converter abstractions.Converter) {
		converter.Destroy()
	})
}
