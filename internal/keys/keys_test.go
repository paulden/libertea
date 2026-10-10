package keys

import "testing"

func TestUnknownLayout(t *testing.T) {
	if _, err := Get("dvorak"); err == nil {
		t.Error("unknown layout should return an error")
	}
	if _, err := Get("ZQSD"); err != nil {
		t.Errorf("layout names should be case insensitive: %v", err)
	}
}
