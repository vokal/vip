package fetch

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
)

func needsRotation(src io.Reader) (bool, int) {
	metadata, err := exif.Decode(src)
	if err != nil {
		fmt.Println(err.Error())
		return false, 0
	}

	orientation, err := metadata.Get(exif.Orientation)
	if err != nil {
		fmt.Println(err.Error())
		return false, 0
	}

	angle := 0
	rotate := false

	switch orientation.String() {
	case "6":
		angle = 90
		rotate = true
	case "3":
		angle = 180
		rotate = true
	case "8":
		angle = 270
		rotate = true
	}

	return rotate, angle
}

func Resize(src io.Reader, c *CacheContext) (io.Reader, error) {
	raw, err := ioutil.ReadAll(src)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	data := bytes.NewReader(raw)

	image, format, err := image.Decode(data)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	buf := new(bytes.Buffer)

	factor := float64(c.Width) / float64(image.Bounds().Size().X)
	height := int(float64(image.Bounds().Size().Y) * factor)

	image = imaging.Resize(image, c.Width, height, imaging.Linear)

	data.Seek(0, 0)

	if rotate, angle := needsRotation(data); rotate {
		switch angle {
		case 90:
			image = imaging.Rotate90(image)
		case 180:
			image = imaging.Rotate180(image)
		case 270:
			image = imaging.Rotate270(image)
		}
	}

	switch format {
	case "jpeg":
		jpeg.Encode(buf, image, nil)
	case "png":
		err = png.Encode(buf, image)
	}

	return buf, err
}

func CenterCrop(src io.Reader, c *CacheContext) (io.Reader, error) {
	raw, err := ioutil.ReadAll(src)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	data := bytes.NewReader(raw)

	image, format, err := image.Decode(data)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	data.Seek(0, 0)

	if rotate, angle := needsRotation(data); rotate {
		switch angle {
		case 90:
			image = imaging.Rotate90(image)
		case 180:
			image = imaging.Rotate180(image)
		case 270:
			image = imaging.Rotate270(image)
		}
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
