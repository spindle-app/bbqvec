package bbq

import (
	"flag"
	"fmt"
	"testing"
)

var (
	nVectors = flag.Int("nvectors", 100000, "Number of vectors to generate")
	testvecs = flag.Int("testvectors", 1000, "Number of vectors to generate")
	dim      = flag.Int("dim", 256, "Dimension of generated vectors")
	nBasis   = flag.Int("bases", 30, "Number of basis sets")
)

func BenchmarkParameters(b *testing.B) {
	//First, build the thing
	vecs := buildVectors(*nVectors, *dim, nil)

	be := NewMemoryBackend(*dim, *nBasis)
	store, err := NewVectorStore(be)
	if err != nil {
		b.Fatal(err)
	}

	for i, v := range vecs {
		store.AddVector(ID(i), v)
	}

	store.BuildIndex()
	targetvecs := buildVectors(*testvecs, *dim, nil)
	res := make([]*ResultSet, *testvecs)
	for i, v := range targetvecs {
		res[i], err = FullTableScanSearch(be, v, 20)
		if err != nil {
			b.Fatal(err)
		}
	}
	for _, searchk := range []int{100, 1000, 10000, 20000} {
		for _, spill := range []int{1, 4, 16, 64} {
			b.Run(fmt.Sprintf("sk%d_sp%d", searchk, spill), func(b *testing.B) {
				benchQuality(b, searchk, spill, store, targetvecs, res)
			})
		}
	}
}

func runBenchmark(b *testing.B, searchk, spill int, store *VectorStore) {
	v := NewRandVector(*dim, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.FindNearest(v, 20, searchk, spill)
	}
}

func benchQuality(b *testing.B, searchk, spill int, store *VectorStore, vecs []Vector, res []*ResultSet) {
	b.ResetTimer()
	//b.ReportAllocs()
	recalls := make([]float64, 4)
	ats := []int{1, 5, 10, 20}
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		v := vecs[i%len(vecs)]
		aknn, err := store.FindNearest(v, 20, searchk, spill)
		if err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
		recalls[0] += aknn.ComputeRecall(res[i%len(vecs)], 1)
		recalls[1] += aknn.ComputeRecall(res[i%len(vecs)], 5)
		recalls[2] += aknn.ComputeRecall(res[i%len(vecs)], 10)
		recalls[3] += aknn.ComputeRecall(res[i%len(vecs)], 20)
	}
	for i, total := range recalls {
		b.ReportMetric(total/float64(b.N), fmt.Sprintf("recall@%02d", ats[i]))
	}
}