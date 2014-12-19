package fetch

import (
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"os"
	"testing"
)

func TestGetMaxWidth(t *testing.T) {
	os.Setenv("VIP_MAX_WIDTH", "500")
	if width := getMaxWidth(); width != 500 {
		t.Fail()
	}

	os.Setenv("VIP_MAX_WIDTH", "1024")
	if width := getMaxWidth(); width != 1024 {
		t.Fail()
	}

	// Test default value
	os.Setenv("VIP_MAX_WIDTH", "")
	if width := getMaxWidth(); width != 720 {
		t.Fail()
	}
}

func TestNeedsRotation(t *testing.T) {
	for i := 1; i <= 8; i++ {
		filename := fmt.Sprintf("f%d-exif.jpg", i)
		f, err := os.Open(fmt.Sprintf("../test/%s", filename))
		if err != nil {
			t.Errorf("Could not open %s.", filename)
		}

		rotate, angle := needsRotation(f)

		switch i {
		case 6:
			if angle != 90 && rotate != true {
				t.Errorf("Expected true, 90; got %d, %t", rotate, angle)
			}
		case 3:
			if angle != 180 && rotate != true {
				t.Errorf("Expected true, 180; got %d, %t", rotate, angle)
			}
		case 8:
			if angle != 270 && rotate != true {
				t.Errorf("Expected true, 270; got %d, %t", rotate, angle)
			}
		default:
			if angle != 0 && rotate != false {
				t.Errorf("Expected false, 0; got %d, %t", rotate, angle)
			}
		}
	}
}

func TestNeedsRotationAltFiles(t *testing.T) {
	filenames := map[int]string{
		1: "awesome.jpeg",
		2: "exif_test_img.jpg",
	}

	for key, filename := range filenames {
		f, err := os.Open(fmt.Sprintf("../test/%s", filename))
		if err != nil {
			t.Errorf("Could not open %s.", filename)
		}

		rotate, angle := needsRotation(f)

		switch key {
		case 1:
			if angle != 0 && rotate != false {
				t.Errorf("Expected true, 90; got %d, %t", rotate, angle)
			}
		case 2:
			if angle != 90 && rotate != true {
				t.Errorf("Expected true, 90; got %d, %t", rotate, angle)
			}
		}
	}
}

func TestUpsideDownImage(t *testing.T) {
	filename := "IMG_0562.JPG"

	f, err := os.Open(fmt.Sprintf("../test/%s", filename))
	if err != nil {
		t.Errorf("Could not open %s because %s", filename, err.Error())
	}

	metadata, err := exif.Decode(f)
	if err != nil {
		t.Errorf("Could not decode EXIF data: %s.", err.Error())
	}

	orientation, err := metadata.Get(exif.Orientation)
	if err != nil {
		t.Errorf("Could not read Orientation: %s.", err.Error())
	}

	t.Logf("The orientation returned is %v", orientation)

}
