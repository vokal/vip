package fetch

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
)

func Round(x float64) float64 {
	var rounder float64
	pow := math.Pow(10, float64(2))
	intermed := x * pow
	_, frac := math.Modf(intermed)
	if frac >= 0.5 {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / pow
}

func Resize(src io.Reader, c *CacheContext) (io.Reader, error) {
	image, format, err := image.Decode(src)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	buf := new(bytes.Buffer)

	factor := float64(c.Width) / float64(image.Bounds().Size().X)
	height := int(float64(image.Bounds().Size().Y) * factor)

	image = imaging.Resize(image, c.Width, height, imaging.Linear)

	switch format {
	case "jpeg":
		jpeg.Encode(buf, image, nil)
	case "png":
		err = png.Encode(buf, image)
	}

	return buf, err
}

func CenterCrop(src io.Reader, c *CacheContext) (io.Reader, error) {
	image, format, err := image.Decode(src)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	buf := new(bytes.Buffer)

	height := image.Bounds().Size().Y
	width := image.Bounds().Size().X

	if width < height {
		image = imaging.CropCenter(image, width, width)
	} else if width > height {
		image = imaging.CropCenter(image, height, height)
	} else {
		image = imaging.CropCenter(image, width, height)
	}

	switch format {
	case "jpeg":
		jpeg.Encode(buf, image, nil)
	case "png":
		err = png.Encode(buf, image)
	}

	return buf, err
}
