package httpimproc

import (
	"fmt"
	"net/url"
	"testing"

	improc "github.com/ourstudio-se/go-image-processor"

	"github.com/stretchr/testify/assert"
)

func Test_That_GetImageSource_Returns_QueryString_Param_URL(t *testing.T) {
	expected := "https://www.test.com/path"
	u, _ := url.Parse(fmt.Sprintf("https://www.test.com/path?url=%s", url.QueryEscape(expected)))
	actual, _ := getImageSource(u.Query())

	assert.Equal(t, expected, actual.String())
}

func Test_That_GetImageSource_Returns_Error_On_Missing_QueryString_Param_URL(t *testing.T) {
	u, _ := url.Parse("https://www.test.com/path?")
	_, err := getImageSource(u.Query())

	assert.Error(t, err)
}

func Test_That_GetImageSource_Returns_Error_On_Non_Absolute_QueryString_Param_URL(t *testing.T) {
	u, _ := url.Parse(fmt.Sprintf("https://www.test.com/path?url=%s", url.QueryEscape("/not/absolute")))
	_, err := getImageSource(u.Query())

	assert.Error(t, err)
}

func Test_That_GetImageSource_Returns_Error_On_Non_HTTP_HTTPS_Scheme_For_QueryString_Param_URL(t *testing.T) {
	u, _ := url.Parse(fmt.Sprintf("https://www.test.com/path?url=%s", url.QueryEscape("ftps://domain")))
	_, err := getImageSource(u.Query())

	assert.Error(t, err)
}

func Test_That_GetFormatSpec_Returns_OutputSpec_Matching_QueryString_Param_Spec(t *testing.T) {
	w := 100
	h := 200
	ax := -1
	ay := 1
	spec := fmt.Sprintf("%dx%d@%d,%d", w, h, ax, ay)
	u, _ := url.Parse(fmt.Sprintf("https://www.test.com/path?spec=%s", spec))
	r, _ := getFormatSpec(u.Query())

	assert.Equal(t, w, int(r.Width))
	assert.Equal(t, h, int(r.Height))
	assert.Equal(t, improc.GravityPull, r.Anchor.Horizontal)
	assert.Equal(t, improc.GravityPush, r.Anchor.Vertical)
}

func Test_That_GetFormatSpec_Returns_OutputSpec_Width_Matching_QueryString_Param_Width(t *testing.T) {
	w := 100
	u, _ := url.Parse(fmt.Sprintf("https://www.test.com/path?width=%d", w))
	r, _ := getFormatSpec(u.Query())

	assert.Equal(t, w, int(r.Width))
}

func Test_That_GetFormatSpec_Returns_OutputSpec_Height_Matching_QueryString_Param_Height(t *testing.T) {
	h := 200
	u, _ := url.Parse(fmt.Sprintf("https://www.test.com/path?height=%d", h))
	r, _ := getFormatSpec(u.Query())

	assert.Equal(t, h, int(r.Height))
}

func Test_That_GetFormatSpec_Returns_Error_When_QueryString_Missing_Dimensions(t *testing.T) {
	u, _ := url.Parse("https://www.test.com/path")
	_, err := getFormatSpec(u.Query())

	assert.Error(t, err)
}

func Test_That_GetFormatSpec_Returns_OutputSpec_Anchors_Matching_QueryString_Params(t *testing.T) {
	ax := -1
	ay := 1
	u, _ := url.Parse(fmt.Sprintf("https://www.test.com/path?width=100&anchorx=%d&anchory=%d", ax, ay))
	r, _ := getFormatSpec(u.Query())

	assert.Equal(t, improc.GravityPull, r.Anchor.Horizontal)
	assert.Equal(t, improc.GravityPush, r.Anchor.Vertical)
}

func Test_That_GetFormatSpec_Returns_Error_On_Missing_QueryString_Param_AnchorX(t *testing.T) {
	u, _ := url.Parse("https://www.test.com/path?width=100&anchory=-1")
	_, err := getFormatSpec(u.Query())

	assert.Error(t, err)
}

func Test_That_GetFormatSpec_Returns_Error_On_Missing_QueryString_Param_AnchorY(t *testing.T) {
	u, _ := url.Parse("https://www.test.com/path?width=100&anchorx=-1")
	_, err := getFormatSpec(u.Query())

	assert.Error(t, err)
}

func Test_GetCompression(t *testing.T) {
	compressions := []struct {
		in  string
		out improc.Compression
	}{
		{"jpg", improc.Jpeg},
		{"JPG", improc.Jpeg},
		{"jpeg", improc.Jpeg},
		{"png", improc.Png},
		{"Png", improc.Png},
		{"webp", improc.WebP},
		{"WEBp", improc.WebP},
		{"notfound", improc.TransitiveCompression},
	}

	for _, tt := range compressions {
		t.Run(tt.in, func(t *testing.T) {
			u, _ := url.Parse(fmt.Sprintf("https://www.test.com/path?out=%s", tt.in))
			r := getCompression(u.Query())

			assert.Equal(t, tt.out, r)
		})
	}
}

func Test_GetBackgroundColor(t *testing.T) {
	colors := []struct {
		in          string
		compression improc.Compression
		out         improc.Color
	}{
		{"", improc.Png, improc.ColorTransparent},
		{"", improc.Jpeg, improc.Color("#FFFFFF")},
		{"abc123", improc.Png, improc.Color("#abc123")},
		{"123abc", improc.Jpeg, improc.Color("#123abc")},
		{"aaa111", improc.WebP, improc.Color("#aaa111")},
	}

	for _, tt := range colors {
		t.Run(tt.in, func(t *testing.T) {
			u, _ := url.Parse(fmt.Sprintf("https://www.test.com/path?background=%s", tt.in))
			r := getBackgroundColor(u.Query(), tt.compression)

			assert.Equal(t, tt.out, r)
		})
	}
}
