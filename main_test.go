package main

import "testing"

func TestGameRestorePick(t *testing.T) {
	s1 := &Stone{}
	s2 := &Stone{}
	s3 := &Stone{}

	type Case struct {
		t        string
		prevNext []*Stone
		prevPick []*Stone
		wantNext []*Stone
		wantPick []*Stone
	}
	cases := []Case{
		Case{
			t:        "as it is",
			prevNext: []*Stone{s1, s2, s3},
			prevPick: []*Stone{},
			wantNext: []*Stone{s1, s2, s3},
			wantPick: []*Stone{},
		},
		Case{
			t:        "restore picked as it is",
			prevNext: []*Stone{},
			prevPick: []*Stone{s1, s2, s3},
			wantNext: []*Stone{s1, s2, s3},
			wantPick: []*Stone{},
		},
		Case{
			t:        "restore in order of pick, next",
			prevNext: []*Stone{s2, s3},
			prevPick: []*Stone{s1},
			wantNext: []*Stone{s1, s2, s3},
			wantPick: []*Stone{},
		},
	}

	for _, cs := range cases {
		g := &Game{}
		g.Pick = cs.prevPick
		g.Next = cs.prevNext
		g.RestorePick()
		if len(g.Pick) != len(cs.wantPick) {
			t.Fatal(cs.t, "pick item num mismatch")
		}
		if len(g.Next) != len(cs.wantNext) {
			t.Fatal(cs.t, "next item num mismatch")
		}

		for n := 0; n < len(g.Pick); n++ {
			if g.Pick[n] != cs.wantPick[n] {
				t.Error(cs.t, "pick item mismatch", n, &g.Pick[n], &cs.wantPick[n])
			}
		}
		for n := 0; n < len(g.Next); n++ {
			if g.Next[n] != cs.wantNext[n] {
				t.Error(cs.t, "next item mismatch", n, &g.Next[n], &cs.wantNext[n])
			}
		}
	}
}

func TestBoardMarkErase(t *testing.T) {
	type C struct {
		x, y int
		c    Color
		e    bool
	}
	type Case struct {
		t string
		c []C
	}
	cases := []Case{
		Case{
			"horizontal",
			[]C{
				C{1, 1, Red, true},
				C{2, 1, Red, true},
				C{3, 1, Red, true},
				C{4, 1, Green, false},
			},
		},
		Case{
			"horizontal not",
			[]C{
				C{1, 1, Red, false},
				C{2, 1, Red, false},
				C{3, 1, Green, false},
				C{4, 1, Red, false},
			},
		},
		Case{
			"vertical",
			[]C{
				C{1, 1, Red, true},
				C{1, 2, Red, true},
				C{1, 3, Red, true},
				C{1, 4, Green, false},
			},
		},

		Case{
			"vertical not",
			[]C{
				C{1, 1, Red, false},
				C{1, 2, Red, false},
				C{1, 3, Green, false},
				C{1, 4, Red, false},
			},
		},

		Case{
			"cross right down",
			[]C{
				C{1, 1, Red, true},
				C{2, 2, Red, true},
				C{3, 3, Red, true},
				C{4, 4, Green, false},
			},
		},

		Case{
			"cross right down 2",
			[]C{
				C{3, 1, Red, true},
				C{4, 2, Red, true},
				C{5, 3, Red, true},
			},
		},

		Case{
			"cross right up",
			[]C{
				C{1, 4, Red, true},
				C{2, 3, Red, true},
				C{3, 2, Red, true},
				C{4, 1, Green, false},
			},
		},

		Case{
			"cross right up 2",
			[]C{
				C{3, 9, Red, true},
				C{4, 8, Red, true},
				C{5, 7, Red, true},
			},
		},
	}
	for _, cs := range cases {
		b := NewBoard()
		for _, c := range cs.c {
			if cell, ok := b.At(c.x, c.y); ok {
				(*cell) = &Stone{Color: c.c}
			} else {
				t.Error(cs.t, "at fail", c.x, c.y)
			}
		}
		b.MarkErase()
		for _, c := range cs.c {
			if cell, ok := b.At(c.x, c.y); ok {
				if (*cell).Erased != c.e {
					t.Error(cs.t, "erased mismatch", c.x, c.y, c.e, (*cell).Erased)
				}
			} else {
				t.Error(cs.t, "at fail", c.x, c.y)
			}
		}
	}
}
