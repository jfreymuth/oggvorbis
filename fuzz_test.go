package oggvorbis

import (
	"bytes"
	"testing"
)

func TestFuzzCrashers(t *testing.T) {
	testData := []string{
		"\xff\xff\xff\xff\xff\xff\xc9\x03",
		// Ogg page header with zero segments: PageSegments-1 underflows
		// the uint8 and indexed an empty segment table.
		"OggS\x00000000000000000000000\x00",
	}

	for _, s := range testData {
		b := bytes.NewReader([]byte(s))
		_, _ = NewReader(b)
	}
}
