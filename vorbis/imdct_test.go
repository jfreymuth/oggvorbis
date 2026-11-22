package vorbis

import (
	"math"
	"math/rand"
	"testing"
)

func imdctSlow(in []float32) []float32 {
	n := len(in)
	fn := float32(n)
	out := make([]float32, 2*n)
	for i := range out {
		var sum float32
		fi := float32(i)
		for k := 0; k < n; k++ {
			fk := float32(k)
			sum += in[k] * float32(math.Cos(float64((math.Pi/fn)*(fi+.5+fn/2)*(fk+.5))))
		}
		out[i] = sum
	}
	return out
}

func TestIMDCT(t *testing.T) {
	//rand.Seed(0)
	const blocksize = 256
	data := make([]float32, blocksize/2)
	for i := range data {
		data[i] = rand.Float32()
	}

	reference := imdctSlow(data)

	var lookup imdctLookup
	generateIMDCTLookup(blocksize, &lookup)
	result := make([]float32, blocksize)
	imdct(&lookup, data, result)

	for i := range result {
		if !equalAbs(result[i], reference[i], 0.001) {
			t.Errorf("different values at index %d (%g != %g)", i, result[i], reference[i])
			break
		}
	}
}

func BenchmarkIMDCT(b *testing.B) {
	const blocksize = 8192

	data := make([]float32, blocksize/2)
	for i := range data {
		data[i] = rand.Float32()
	}

	var lookup imdctLookup
	generateIMDCTLookup(blocksize, &lookup)
	out := make([]float32, blocksize)

	in := make([]float32, blocksize/2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// the imdct function does not preserve the input,
		// reset it, otherwise everything ends up being NaN
		copy(in, data)
		imdct(&lookup, in, out)
	}
}

func equalAbs(a, b, tolerance float32) bool {
	return a-b < tolerance && b-a < tolerance
}
