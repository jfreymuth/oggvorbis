package vorbis

import (
	"encoding/binary"
	"encoding/gob"
	"io"
	"os"
	"testing"
)

type GobVorbis struct {
	Headers [3][]byte
	Packets [][]byte
}

func readTestFile() (*GobVorbis, error) {
	file, err := os.Open("testdata/test.gob")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := new(GobVorbis)
	err = gob.NewDecoder(file).Decode(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func TestDecode(t *testing.T) {
	data, err := readTestFile()
	if err != nil {
		t.Fatal(err)
	}

	var dec Decoder
	for _, header := range data.Headers {
		err = dec.ReadHeader(header)
		if err != nil {
			t.Fatal(err)
		}
	}

	if dec.SampleRate() != 44100 {
		t.Errorf("sample rate is %d, expected %d", dec.SampleRate(), 44100)
	}
	if dec.Channels() != 1 {
		t.Errorf("channels is %d, expected %d", dec.Channels(), 1)
	}

	refFile, err := os.Open("testdata/test.raw")
	if err != nil {
		t.Fatal(err)
	}
	defer refFile.Close()

	rawSize, _ := refFile.Seek(0, io.SeekEnd)
	refFile.Seek(0, io.SeekStart)
	reference := make([]float32, rawSize/4)
	binary.Read(refFile, binary.LittleEndian, reference)

	pos := 0
decode:
	for _, packet := range data.Packets {
		out, err := dec.Decode(packet)
		if err != nil {
			t.Fatal(err)
		}
		n := len(out)
		if pos+n > len(reference) {
			// it's ok for the decoded data to be longer, since only an integral number of packets can be decoded
			n = len(reference) - pos
		}
		for i := 0; i < n; i++ {
			if !equal(out[i], reference[pos+i], .00002) {
				t.Errorf("different values at sample %d (%g != %g)", i, out[i], reference[pos+i])
				break decode
			}
		}
		pos += n
	}
	if pos < len(reference) {
		t.Errorf("not enough samples were decoded (%d, expected at least %d)", pos, len(reference))
	}
}

func equal(a, b, tolerance float32) bool {
	return (a > b && a-b < tolerance) || b-a < tolerance
}

func BenchmarkSetup(b *testing.B) {
	data, err := readTestFile()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var dec Decoder
		for _, header := range data.Headers {
			if err = dec.ReadHeader(header); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	data, err := readTestFile()
	if err != nil {
		b.Fatal(err)
	}

	var dec Decoder
	for _, header := range data.Headers {
		if err = dec.ReadHeader(header); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, packet := range data.Packets {
			if _, err := dec.Decode(packet); err != nil {
				b.Fatal(err)
			}
		}
	}
}
