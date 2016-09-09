package fetch

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"math"

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

	width := c.Width
	data := bytes.NewReader(raw)
	img, format, err := image.Decode(data)
	if err != nil {
		return nil, err
	}
	var resizedImage image.NRGBA
	if c.Crop {

		minDimension := int(math.Min(float64(img.Bounds().Size().X), float64(img.Bounds().Size().Y)))

		if minDimension < c.Width || c.Width == 0 {
			width = minDimension
		}

		resizedImage = *imaging.Fill(img, width, width, imaging.Center, imaging.Lanczos)
	} else {
		resizedImage = *imaging.Resize(img, width, 0, imaging.Lanczos)
	}

	buf := new(bytes.Buffer)
	var imgFormat imaging.Format
	switch format {
	case "png":
		imgFormat = imaging.PNG
	case "jpeg":
		imgFormat = imaging.JPEG
	case "tiff":
		imgFormat = imaging.TIFF
	case "bmp":
		imgFormat = imaging.BMP
	default:
		return nil, errors.New("unsupported image format")
	}

	err = imaging.Encode(buf, resizedImage.SubImage(resizedImage.Rect), imgFormat)
	if err != nil {
		return nil, err
	}

	return buf, err

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
