package main

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestOutingString(t *testing.T) {
	o := outing([]group{
		{"brad", "mom", "dad"},
		{"ryan", "cole", "emma"},
	})
	t.Logf("String is: %s", o.String())
}

func TestRandOuting(t *testing.T) {
	ps := []playerID{"a", "b", "c", "d", "e", "f", "g"}
	sizes := sizesOf(len(ps))
	for i := 0; i < 10; i++ {
		o := randOuting(ps, sizes)
		t.Logf("Outing: %s", o)
	}
}

func TestSizes(t *testing.T) {
	cases := []struct {
		in   int
		want []int
	}{
		{6, []int{3, 3}},
		{16, []int{4, 4, 4, 4}},
		{10, []int{3, 3, 4}},
	}
	for _, tc := range cases {
		got := sizesOf(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("sizes(%d) = %v; want %v", tc.in, got, tc.want)
		}
	}
}

func TestForeachPair(t *testing.T) {
	foreachPair([]playerID{"c", "a", "t", "s"}, func(p pair) {
		t.Logf("%v", p)
	})
}

func TestScoring(t *testing.T) {
	memHistory = nil
	oldAdjust := pairAdjust
	pairAdjust = nil
	defer func() { pairAdjust = oldAdjust }()
	players := []playerID{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
		"u", "v", "w", "x", "y", "z", "A", "B",
	}
	var nGames = 100
	if testing.Short() {
		nGames = 10
	}
	for i := 0; i < nGames; i++ {
		hist, err := loadHistory(nil)
		if err != nil {
			t.Fatal(err)
		}
		hs := hist.summarize()
		os := makeGroups(hs, players)
		if len(os) == 0 {
			t.Fatal("no outings")
		}
		for _, g := range os[0].o {
			addGroupPlayHistory(nil, groupPlay{
				Players: g,
				When:    time.Now(),
			})
		}
	}
	hist, err := loadHistory(nil)
	if err != nil {
		t.Fatal(err)
	}
	hs := hist.summarize()

	dist := make([]int, nGames)
	for p, ct := range hs.count {
		t.Logf("%v %d", p, ct)
		dist[ct]++
	}
	for i, n := range dist {
		if n != 0 {
			t.Logf("%4d: %4d %s", i, n, strings.Repeat("x", n))
		}
	}
}
