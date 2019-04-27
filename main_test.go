package main

import "testing"

func TestBoardMarkErase(t *testing.T) {
	b := NewBoard()
	b.Cell[1][1] = &Stone{Color: Red}
	b.Cell[2][1] = &Stone{Color: Red}
	b.Cell[3][1] = &Stone{Color: Red}
	b.Cell[4][1] = &Stone{Color: Green}
	b.MarkErase()
	if !b.Cell[1][1].Erased {
		t.Fail()
	}
	if !b.Cell[2][1].Erased {
		t.Fail()
	}
	if !b.Cell[3][1].Erased {
		t.Fail()
	}
	if b.Cell[4][1].Erased {
		t.Fail()
	}

	b2 := NewBoard()
	b2.Cell[1][1] = &Stone{Color: Red}
	b2.Cell[2][1] = &Stone{Color: Red}
	b2.Cell[3][1] = &Stone{Color: Green}
	b2.Cell[4][1] = &Stone{Color: Red}
	b2.MarkErase()
	if b2.Cell[1][1].Erased {
		t.Error(1, 1)
	}
	if b2.Cell[2][1].Erased {
		t.Error(2, 1)
	}
	if b2.Cell[3][1].Erased {
		t.Error(3, 1)
	}
	if b2.Cell[4][1].Erased {
		t.Error(4, 1)
	}
}
