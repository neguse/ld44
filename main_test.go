package main

import (
	"testing"
)

func TestCalcScore(t *testing.T) {
	type Case struct {
		t       string
		sequent int
		num     int
		ans     int
		equ     string
	}
	cases := []Case{
		Case{
			"1",
			1, 3,
			6,
			"2x3=6.",
		},
	}
	for _, cs := range cases {
		ans, equ := CalcScore(cs.sequent, cs.num)
		if ans != cs.ans || equ != cs.equ {
			t.Error(cs.t, ans, cs.ans, equ, cs.equ)
		}
	}
}

func TestGameHeight(t *testing.T) {
	g := NewGame()
	if c, ok := g.Board.At(1, 1); ok {
		(*c) = &Stone{Color: Red}
	} else {
		t.FailNow()
	}

	height := g.Board.HeightAt(1)
	if height != 1 {
		t.Error("height", 1, height)
	}
}

func TestGameAdjustPick(t *testing.T) {
	type C struct {
		x, y int
		c    Color
	}
	type P struct {
		t            string
		cx, cy       int
		px, py, plen int
	}
	type Case struct {
		t string
		c []C
		p []P
	}
	cases := []Case{
		Case{
			"emptywall",
			[]C{},
			[]P{
				P{"-", 0, 0, 1, 0, 0},
			},
		},
		Case{
			"empty",
			[]C{},
			[]P{
				P{"-", 1, PickMax - 1, 1, PickMax, 1},
			},
		},
		Case{
			"just+1",
			[]C{
				C{1, PickMax + 1, Red},
			},
			[]P{
				P{"+1", 1, PickMax, 1, PickMax - 1, 0},
				P{" 0", 1, PickMax - 1, 1, PickMax - 1, 1},
				P{"-1", 1, PickMax - 2, 1, PickMax - 1, 2},
			},
		},
		Case{
			"just",
			[]C{
				C{1, PickMax, Red},
			},
			[]P{
				P{"+1", 1, PickMax, 1, PickMax - 2, 0},
				P{" 0", 1, PickMax - 1, 1, PickMax - 2, 0},
				P{"-1", 1, PickMax - 2, 1, PickMax - 2, 1},
				P{"-2", 1, PickMax - 3, 1, PickMax - 2, 2},
			},
		},
		Case{
			"just-1",
			[]C{
				C{1, PickMax - 1, Red},
			},
			[]P{
				P{"+1", 1, PickMax, 1, PickMax - 3, 0},
				P{" 0", 1, PickMax - 1, 1, PickMax - 3, 0},
				P{"-1", 1, PickMax - 2, 1, PickMax - 3, 0},
				P{"-2", 1, PickMax - 3, 1, PickMax - 3, 1},
				P{"-3", 1, PickMax - 4, 1, PickMax - 3, 2},
			},
		},
		Case{
			"full+1",
			[]C{
				C{1, 2, Red},
			},
			[]P{
				P{"+1", 1, 2, 1, 0, 0},
				P{" 0", 1, 1, 1, 0, 0},
				P{"-1", 1, 0, 1, 0, 1},
			},
		},
		Case{
			"full",
			[]C{
				C{1, 1, Red},
			},
			[]P{
				P{"+1", 1, 1, 1, -1, 0},
				P{" 0", 1, 0, 1, -1, 0},
				P{"-1", 1, -1, 1, -1, 0},
			},
		},
	}
	for _, cs := range cases {
		g := NewGame()
		for _, p := range cs.p {
			for _, cell := range cs.c {
				if c, ok := g.Board.At(cell.x, cell.y); ok {
					(*c) = &Stone{Color: cell.c}
				}
				g.AdjustPick(p.cx, p.cy)
				if g.PickX != p.px || g.PickY != p.py || g.PickLen != p.plen {
					t.Error(cs.t, p.t, p.px, g.PickX, p.py, g.PickY, p.plen, g.PickLen)
				}
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

		Case{
			"jammer",
			[]C{
				C{1, 3, Red, true},
				C{2, 3, Red, true},
				C{3, 3, Red, true},
				C{4, 3, Jammer, true},
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
