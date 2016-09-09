package fetch

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"math"

	"github.com/daddye/vips"
	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

func needsRotation(src io.Reader) int {
	metadata, err := exif.Decode(src)
	if err != nil {
		return 0
	}

	orientation, err := metadata.Get(exif.Orientation)
	if err != nil {
		return 0
	}

	switch orientation.String() {
	case "6":
		return 270
	case "3":
		return 180
	case "8":
		return 90
	default:
		return 0
	}

}

func GetRotatedImage(src io.Reader) (image.Image, string, error) {
	raw, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, "", err
	}

	data := bytes.NewReader(raw)

	image, format, err := image.Decode(data)
	if err != nil {
		return nil, "", err
	}

	if _, err := data.Seek(0, 0); err != nil {
		return nil, "", err
	}

	angle := needsRotation(data)
	switch angle {
	case 90:
		image = imaging.Rotate90(image)
	case 180:
		image = imaging.Rotate180(image)
	case 270:
		image = imaging.Rotate270(image)
	}

	return image, format, nil
}

func Resize(src io.Reader, c *CacheContext) (io.Reader, error) {
	raw, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	options := vips.Options{
		Width:        c.Width,
		Crop:         true,
		Extend:       vips.EXTEND_WHITE,
		Interpolator: vips.BILINEAR,
		Gravity:      vips.CENTRE,
		Quality:      80,
	}

	if c.Crop {
		data := bytes.NewReader(raw)

		image, _, err := image.Decode(data)
		if err != nil {
			return nil, err
		}

		minDimension := int(math.Min(float64(image.Bounds().Size().X), float64(image.Bounds().Size().Y)))

		if minDimension < options.Width || options.Width == 0 {
			options.Width = minDimension
		}

		options.Height = options.Width
	}

	res, err := vips.Resize(raw, options)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(res), err
}

func ResizeGif(src io.Reader, c *CacheContext) (io.Reader, error) {
	raw, format, err := image.Decode(src)
	if err != nil {
		return nil, err
	}
	if format != "gif" {
		return nil, errors.New("Aborted attempt to resize another type as a gif")
	}

	pngBuf := new(bytes.Buffer)
	if png.Encode(pngBuf, raw) != nil {
		return nil, err
	}

	return Resize(pngBuf, c)
}
