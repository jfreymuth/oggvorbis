package oggvorbis_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"testing"

	"github.com/jfreymuth/oggvorbis"
)

type reader struct {
	io.Reader
}

func (r reader) Read(p []byte) (int, error) {
	return r.Reader.Read(p)
}

func test(t *testing.T, filename string, reference []float32, tolerance float32, wrap bool) {
	ogg, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer ogg.Close()
	var r io.Reader
	if wrap {
		// wrap the file so it doesn't implement io.Seeker
		r = reader{ogg}
	} else {
		r = ogg
	}
	dec, _, err := oggvorbis.ReadAll(r)
	if err != nil {
		t.Fatal("decoding error: ", err)
	}

	compare(t, dec, reference, tolerance)
}

func TestRead(t *testing.T) {
	raw, err := os.Open("testdata/test.raw")
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()

	rawSize, _ := raw.Seek(0, io.SeekEnd)
	raw.Seek(0, io.SeekStart)
	rawData := make([]float32, rawSize/4)
	binary.Read(raw, binary.LittleEndian, rawData)

	test(t, "testdata/test.ogg", rawData, 0.00002, false)
	test(t, "testdata/test.ogg", rawData, 0.00002, true)
}

func TestFormat(t *testing.T) {
	ogg, err := os.Open("testdata/test.ogg")
	if err != nil {
		t.Fatal(err)
	}
	defer ogg.Close()

	format, err := oggvorbis.GetFormat(ogg)
	if err != nil {
		t.Fatal(err)
	}
	if format.SampleRate != 44100 {
		t.Errorf("sample rate is %d, expected %d", format.SampleRate, 44100)
	}
	if format.Channels != 1 {
		t.Errorf("channels is %d, expected %d", format.Channels, 1)
	}
}

func TestLength(t *testing.T) {
	ogg, err := os.Open("testdata/test.ogg")
	if err != nil {
		t.Fatal(err)
	}
	defer ogg.Close()

	length, _, err := oggvorbis.GetLength(ogg)
	if err != nil {
		t.Fatal(err)
	}
	if length != 44100 {
		t.Errorf("length is %d, expected %d", length, 44100)
	}
}

func TestLengthForUnexpectedEof(t *testing.T) {
	ogg, err := os.Open("testdata/eof_issue.ogg")
	if err != nil {
		t.Fatal(err)
	}
	defer ogg.Close()

	length, _, err := oggvorbis.GetLength(ogg)
	if err != nil {
		t.Fatal(err)
	}

	const expectedLength = 72384
	if length != expectedLength {
		t.Errorf("length is %d, expected %d", length, expectedLength)
	}
	r, err := oggvorbis.NewReader(ogg)
	if err != nil {
		t.Fatal(err)
	}

	if r.Length() != expectedLength {
		t.Fatalf("length is %d, expected %d", r.Length(), expectedLength)
	}
}

func TestSeek(t *testing.T) {
	ogg, err := os.Open("testdata/long.ogg")
	if err != nil {
		t.Fatal(err)
	}
	defer ogg.Close()

	data, _, err := oggvorbis.ReadAll(ogg)
	if err != nil {
		t.Fatal(err)
	}
	ogg.Seek(0, io.SeekStart)

	r, err := oggvorbis.NewReader(ogg)
	if err != nil {
		t.Fatal(err)
	}

	if int(r.Length()) != len(data) {
		t.Fatalf("length is %d, expected %d", r.Length(), len(data))
	}

	sections := []struct{ start, len int64 }{
		{600000, 10000},
		// test first position
		{0, 500},
		// test last position
		{881999, 1},
		// test around page boundary
		{182399, 100},
		{182400, 100},
		{182401, 100},
		// seek to an out of bounds position and then continue
		{1000000, 0},
		{200000, 1000},
	}
test:
	for _, s := range sections {
		err := r.SetPosition(s.start)
		if err != nil {
			t.Error(err)
			continue test
		}
		if s.len == 0 {
			continue test
		}
		ref := data[s.start : s.start+s.len]
		buf := make([]float32, s.len)
		for n := 0; n < len(buf); {
			read, err := r.Read(buf[n:])
			if err != nil {
				t.Error(err)
				continue test
			}
			n += read
		}
		compare(t, buf, ref, .00001)
	}

	r.SetPosition(882000)
	_, err = r.Read(nil)
	if err != io.EOF {
		t.Errorf("error should be EOF, but is %s", err)
	}
}

func TestEmpty(t *testing.T) {
	_, err := oggvorbis.NewReader(bytes.NewReader(nil))
	if err != io.ErrUnexpectedEOF {
		t.Errorf("error should be unexpected EOF, but is %s", err)
	}
}

func compare(t *testing.T, a, e []float32, tolerance float32) {
	if len(a) != len(e) {
		t.Fatalf("length of output is %d, expected %d", len(a), len(e))
	}
	for i := range a {
		if !equal(a[i], e[i], tolerance) {
			t.Errorf("different values at index %d (%g != %g)", i, a[i], e[i])
			break
		}
	}
}

func equal(a, b, tolerance float32) bool {
	return (a > b && a-b < tolerance) || b-a < tolerance
}
