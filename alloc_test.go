package oggvorbis

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// A streaming consumer reads packets for every voice every few milliseconds
// and seeks at every loop wrap, so the Ogg layer must allocate nothing once
// its page buffers have grown to the clip's largest page. These tests pin
// that for the Ogg layer alone and never call the decoder, so they hold
// whatever the vorbis package allocates.

func openAllocClip(t *testing.T) *oggReader {
	t.Helper()
	data, err := os.ReadFile("testdata/test.ogg")
	if err != nil {
		t.Fatal(err)
	}
	source := bytes.NewReader(data)
	return &oggReader{source: source, seeker: source}
}

func TestNextPacketAllocatesNothing(t *testing.T) {
	r := openAllocClip(t)
	readClip := func() {
		r.seeker.Seek(0, io.SeekStart)
		r.buffer = nil
		r.ready = false
		r.lastPacket = false
		r.continued = r.continued[:0]
		packets := 0
		for !r.lastPacket {
			if _, err := r.NextPacket(); err != nil {
				t.Fatal(err)
			}
			packets++
		}
		if packets < 4 {
			t.Fatalf("read %d packets from the clip", packets)
		}
	}
	readClip() // grows the page buffers to the largest page in the clip
	if allocs := testing.AllocsPerRun(10, readClip); allocs != 0 {
		t.Errorf("reading the clip's packets allocates %v times, expected 0", allocs)
	}
}

func TestSeekPageBeforeAllocatesNothing(t *testing.T) {
	r := openAllocClip(t)
	length, err := r.LastPosition()
	if err != nil {
		t.Fatal(err)
	}
	positions := []int64{0, length / 3, length / 2, length - 1, 0}
	seekAll := func() {
		for _, pos := range positions {
			if _, err := r.SeekPageBefore(pos); err != nil {
				t.Fatal(err)
			}
			if _, err := r.NextPacket(); err != nil {
				t.Fatal(err)
			}
		}
	}
	seekAll()
	if allocs := testing.AllocsPerRun(10, seekAll); allocs != 0 {
		t.Errorf("seeking allocates %v times, expected 0", allocs)
	}
}
