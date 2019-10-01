package oggvorbis

import (
	"bytes"
	"testing"
)

func TestFuzzCrashers(t *testing.T) {
	testData := []string{
		"\xff\xff\xff\xff\xff\xff\xc9\x03",
	}

	for _, s := range testData {
		b := bytes.NewReader([]byte(s))
		_, _ = NewReader(b)
	}
}
