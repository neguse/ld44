package main

import "testing"

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
