package improc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_That_GetHorizontalAnchorValue_Returns_Leftest_Pixel_For_GravityPull_For_Downscale(t *testing.T) {
	anchor := &Anchor{
		Horizontal: GravityPull,
	}

	currentWidth := float64(200)
	nextWidth := float64(100)

	result := anchor.GetHorizontalAnchorValue(currentWidth, nextWidth)

	assert.Equal(t, 0, result)
}

func Test_That_GetHorizontalAnchorValue_Returns_Leftest_Pixel_For_GravityPull_For_Upscale(t *testing.T) {
	anchor := &Anchor{
		Horizontal: GravityPull,
	}

	currentWidth := float64(100)
	nextWidth := float64(200)

	result := anchor.GetHorizontalAnchorValue(currentWidth, nextWidth)

	assert.Equal(t, 0, result)
}

func Test_That_GetHorizontalAnchorValue_Returns_Rightest_Pixel_For_GravityPush_For_Downscale(t *testing.T) {
	anchor := &Anchor{
		Horizontal: GravityPush,
	}

	currentWidth := float64(200)
	nextWidth := float64(100)

	result := anchor.GetHorizontalAnchorValue(currentWidth, nextWidth)

	assert.Equal(t, -100, result)
}

func Test_That_GetHorizontalAnchorValue_Returns_Rightest_Pixel_For_GravityPush_For_Upscale(t *testing.T) {
	anchor := &Anchor{
		Horizontal: GravityPush,
	}

	currentWidth := float64(100)
	nextWidth := float64(200)

	result := anchor.GetHorizontalAnchorValue(currentWidth, nextWidth)

	assert.Equal(t, 100, result)
}

func Test_That_GetHorizontalAnchorValue_Returns_Center_Pixel_For_GravityCenter_For_Downscale(t *testing.T) {
	anchor := &Anchor{
		Horizontal: GravityCenter,
	}

	currentWidth := float64(200)
	nextWidth := float64(100)

	result := anchor.GetHorizontalAnchorValue(currentWidth, nextWidth)

	assert.Equal(t, -50, result)
}

func Test_That_GetHorizontalAnchorValue_Returns_Center_Pixel_For_GravityCenter_For_Upscale(t *testing.T) {
	anchor := &Anchor{
		Horizontal: GravityCenter,
	}

	currentWidth := float64(100)
	nextWidth := float64(200)

	result := anchor.GetHorizontalAnchorValue(currentWidth, nextWidth)

	assert.Equal(t, 50, result)
}

func Test_That_GetVerticalAnchorValue_Returns_Leftest_Pixel_For_GravityPull_For_Downscale(t *testing.T) {
	anchor := &Anchor{
		Vertical: GravityPull,
	}

	currentWidth := float64(200)
	nextWidth := float64(100)

	result := anchor.GetVerticalAnchorValue(currentWidth, nextWidth)

	assert.Equal(t, 0, result)
}

func Test_That_GetVerticalAnchorValue_Returns_Leftest_Pixel_For_GravityPull_For_Upscale(t *testing.T) {
	anchor := &Anchor{
		Vertical: GravityPull,
	}

	currentWidth := float64(100)
	nextWidth := float64(200)

	result := anchor.GetVerticalAnchorValue(currentWidth, nextWidth)

	assert.Equal(t, 0, result)
}

func Test_That_GetVerticalAnchorValue_Returns_Rightest_Pixel_For_GravityPush_For_Downscale(t *testing.T) {
	anchor := &Anchor{
		Vertical: GravityPush,
	}

	currentWidth := float64(200)
	nextWidth := float64(100)

	result := anchor.GetVerticalAnchorValue(currentWidth, nextWidth)

	assert.Equal(t, -100, result)
}

func Test_That_GetVerticalAnchorValue_Returns_Rightest_Pixel_For_GravityPush_For_Upscale(t *testing.T) {
	anchor := &Anchor{
		Vertical: GravityPush,
	}

	currentWidth := float64(100)
	nextWidth := float64(200)

	result := anchor.GetVerticalAnchorValue(currentWidth, nextWidth)

	assert.Equal(t, 100, result)
}

func Test_That_GetVerticalAnchorValue_Returns_Center_Pixel_For_GravityCenter_For_Downscale(t *testing.T) {
	anchor := &Anchor{
		Vertical: GravityCenter,
	}

	currentWidth := float64(200)
	nextWidth := float64(100)

	result := anchor.GetVerticalAnchorValue(currentWidth, nextWidth)

	assert.Equal(t, -50, result)
}

func Test_That_GetVerticalAnchorValue_Returns_Center_Pixel_For_GravityCenter_For_Upscale(t *testing.T) {
	anchor := &Anchor{
		Vertical: GravityCenter,
	}

	currentWidth := float64(100)
	nextWidth := float64(200)

	result := anchor.GetVerticalAnchorValue(currentWidth, nextWidth)

	assert.Equal(t, 50, result)
}

func Test_That_ParseOutputSpec_Defaults_To_Center_Gravity_Anchors(t *testing.T) {
	raw := "200x100"
	spec, _ := ParseOutputSpec(raw)

	assert.Equal(t, GravityCenter, spec.Anchor.Horizontal)
	assert.Equal(t, GravityCenter, spec.Anchor.Vertical)
}

func Test_That_ParseOutputSpec_Returns_Error_On_Missing_Dimension_Separator(t *testing.T) {
	raw := "100"
	_, err := ParseOutputSpec(raw)

	assert.Error(t, err)
}

func Test_That_ParseOutputSpec_Returns_Error_On_Multiple_Dimension_Separators(t *testing.T) {
	raw := "100x100x100"
	_, err := ParseOutputSpec(raw)

	assert.Error(t, err)
}

func Test_That_ParseOutputSpec_Parses_Width_And_Height(t *testing.T) {
	width := 100
	height := 200
	raw := fmt.Sprintf("%dx%d", width, height)
	spec, _ := ParseOutputSpec(raw)

	assert.Equal(t, width, int(spec.Width))
	assert.Equal(t, height, int(spec.Height))
}

func Test_That_ParseOutputSpec_Parses_Only_Width_Dimension(t *testing.T) {
	width := 100
	raw := fmt.Sprintf("%dx", width)
	spec, _ := ParseOutputSpec(raw)

	assert.Equal(t, width, int(spec.Width))
	assert.Equal(t, 0, int(spec.Height))
}

func Test_That_ParseOutputSpec_Parses_Only_Height_Dimension(t *testing.T) {
	height := 200
	raw := fmt.Sprintf("x%d", height)
	spec, _ := ParseOutputSpec(raw)

	assert.Equal(t, height, int(spec.Height))
	assert.Equal(t, 0, int(spec.Width))
}

func Test_That_ParseAnchorSpec_Sets_GravityPull_For_Horizontal_Negative_Value(t *testing.T) {
	raw := "-1,9"
	spec := ParseAnchorSpec(raw)

	assert.Equal(t, GravityPull, spec.Horizontal)
}

func Test_That_ParseAnchorSpec_Sets_GravityCenter_For_Horizontal_Zero_Value(t *testing.T) {
	raw := "0,9"
	spec := ParseAnchorSpec(raw)

	assert.Equal(t, GravityCenter, spec.Horizontal)
}

func Test_That_ParseAnchorSpec_Sets_GravityPush_For_Horizontal_Positive_Value(t *testing.T) {
	raw := "1,9"
	spec := ParseAnchorSpec(raw)

	assert.Equal(t, GravityPush, spec.Horizontal)
}

func Test_That_ParseAnchorSpec_Sets_GravityPull_For_Vertical_Negative_Value(t *testing.T) {
	raw := "9,-1"
	spec := ParseAnchorSpec(raw)

	assert.Equal(t, GravityPull, spec.Vertical)
}

func Test_That_ParseAnchorSpec_Sets_GravityCenter_For_Vertical_Zero_Value(t *testing.T) {
	raw := "9,0"
	spec := ParseAnchorSpec(raw)

	assert.Equal(t, GravityCenter, spec.Vertical)
}

func Test_That_ParseAnchorSpec_Sets_GravityPush_For_Vertical_Positive_Value(t *testing.T) {
	raw := "9,1"
	spec := ParseAnchorSpec(raw)

	assert.Equal(t, GravityPush, spec.Vertical)
}

func Test_That_ParseAnchorSpec_Defaults_To_CenterGravity_Values_For_Separator_Error(t *testing.T) {
	raw := "1"
	spec := ParseAnchorSpec(raw)

	assert.Equal(t, GravityCenter, spec.Horizontal)
	assert.Equal(t, GravityCenter, spec.Vertical)
}

func Test_That_ParseAnchorSpec_Defaults_To_CenterGravity_Horizontal_Value(t *testing.T) {
	raw := "k,1"
	spec := ParseAnchorSpec(raw)

	assert.Equal(t, GravityCenter, spec.Horizontal)
}

func Test_That_ParseAnchorSpec_Defaults_To_CenterGravity_Vertical_Value(t *testing.T) {
	raw := "1,k"
	spec := ParseAnchorSpec(raw)

	assert.Equal(t, GravityCenter, spec.Vertical)
}
