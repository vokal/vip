package fetch

import (
	"fmt"
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

		angle := needsRotation(f)

		switch i {
		case 6:
			if angle != 270 {
				t.Errorf("Expected 270; got %d", angle)
			}
		case 3:
			if angle != 180 {
				t.Errorf("Expected 180; got %d", angle)
			}
		case 8:
			if angle != 90 {
				t.Errorf("Expected 90; got %d", angle)
			}
		default:
			if angle != 0 {
				t.Errorf("Expected 0; got %d", angle)
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

		angle := needsRotation(f)

		switch key {
		case 1:
			if angle != 0 {
				t.Errorf("Expected 0; got %d", angle)
			}
		case 2:
			if angle != 270 {
				t.Errorf("Expected 270; got %d", angle)
			}
		}
	}
}
