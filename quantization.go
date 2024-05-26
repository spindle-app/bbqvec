package bbq

import (
	"encoding/binary"
	"math"

	"github.com/viterin/vek/vek32"
)

type Quantization[L any] interface {
	Similarity(x, y L) float32
	Lower(v Vector) (L, error)
	Marshal(lower L) ([]byte, error)
	Unmarshal(data []byte) (L, error)
}

var _ Quantization[Vector] = NoQuantization{}

type NoQuantization struct{}

func (q NoQuantization) Similarity(x, y Vector) float32 {
	return vek32.CosineSimilarity(x, y)
}

func (q NoQuantization) Lower(v Vector) (Vector, error) {
	return v.Clone(), nil
}

func (q NoQuantization) Marshal(lower Vector) ([]byte, error) {
	out := make([]byte, 4*len(lower))
	for i, n := range lower {
		u := math.Float32bits(n)
		binary.LittleEndian.PutUint32(out[i*4:], u)
	}
	return out, nil
}

func (q NoQuantization) Unmarshal(data []byte) (Vector, error) {
	out := make([]float32, len(data)>>2)
	for i := 0; i < len(data); i += 4 {
		bits := binary.LittleEndian.Uint32(data[i:])
		out[i>>2] = math.Float32frombits(bits)
	}
	return out, nil
}