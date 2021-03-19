package linalg_test

import (
	"math/rand"
	"strconv"
	"testing"

	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/linalg"
)

func TestPageRank(t *testing.T) {
	edgs := map[uint64]map[uint64]int{}
	edgs[0x5B92] = map[uint64]int{}
	edgs[0x804F] = map[uint64]int{}
	edgs[0xFB26] = map[uint64]int{}
	edgs[0xF523] = map[uint64]int{}
	edgs[0xFC63] = map[uint64]int{}
	edgs[0x804F][0x5B92]++
	edgs[0xFB26][0x5B92]++
	edgs[0xFB26][0x804F]++
	edgs[0xF523][0x5B92]++
	edgs[0xF523][0x804F]++
	edgs[0xF523][0xFB26]++
	edgs[0xFC63][0x5B92]++
	edgs[0xFC63][0x804F]++
	edgs[0xFC63][0xFB26]++
	edgs[0xFC63][0xF523]++
	got := linalg.PageRank(edgs)
	want := map[uint64]float64{}
	want[0x5B92] = 0.539773357682638
	want[0x804F] = 0.20997909420705596
	want[0xFB26] = 0.11761540730647063
	want[0xF523] = 0.07719901505201851
	want[0xFC63] = 0.055433125751816706
	if !EqualMap(got, want) {
		t.Error("!equalMap(got, want)")
	}
}

func BenchmarkPageRank(b *testing.B) {
	for i := 995; i <= 1005; i++ {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			edgs := map[uint64]map[uint64]int{}
			for j := 0; j < i; j++ {
				node := uint64(j)
				edgs[node] = map[uint64]int{}
			}
			for j := 0; j < i; j++ {
				from := uint64(rand.Intn(i))
				to := uint64(rand.Intn(i))
				edgs[from][to]++
			}
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				linalg.PageRank(edgs)
			}
		})
	}
}
