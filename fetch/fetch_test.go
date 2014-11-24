package fetch

import (
	"os"
	"testing"
)

func TestgetMaxWidth(t *testing.T) {
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
