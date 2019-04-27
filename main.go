package main

import (
	"image"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"

	_ "github.com/neguse/ld44/statik"
	"github.com/rakyll/statik/fs"
)

type Stone struct {
	Color  Color
	Erased bool
}

type Color int

const (
	None Color = iota
	Red
	Blue
	Green
	Yellow
	Wall
)

type Step int

const (
	Move Step = iota
	FallPick
	FixPick
	Erase
	FallStone
)

var Colors []Color = []Color{
	None,
	Red,
	Blue,
	Green,
	Yellow,
	Wall,
}

const (
	ScreenWidth  = 240
	ScreenHeight = 320

	BoardWidth  = 8
	BoardHeight = 12

	StoneWidth  = 16
	StoneHeight = 16
)

var Texture *ebiten.Image
var StoneImages map[Color]*ebiten.Image
var G *Game

type Game struct {
	Board        *Board
	Pick         []*Stone
	PickX, PickY int
	Step         Step
}

func (g *Game) NewColoredStone() *Stone {
	r := rand.Intn(4)
	return &Stone{
		Color: []Color{Red, Blue, Green, Yellow}[r],
	}
}

func NewWall() *Stone {
	return &Stone{
		Color: Wall,
	}
}

func (g *Game) InitPick() {
	g.PickX = 3
	g.PickY = 1
	g.Pick = []*Stone{G.NewColoredStone(), G.NewColoredStone()}
}

func (g *Game) Update() {
	switch g.Step {
	case Move:
		if inpututil.IsKeyJustPressed(ebiten.KeyH) {
			if 1 < g.PickX {
				g.PickX--
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyL) {
			if g.PickX < BoardWidth-2 {
				g.PickX++
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyK) {
			g.Pick = append(g.Pick, g.NewColoredStone())
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
			g.Step = FallPick
		}
	case FallPick:
		if !g.IsPickCollide(g.PickX, g.PickY+1) {
			g.PickY++
		} else {
			g.Step = FixPick
		}
	case FixPick:
		g.FixPick()
		g.Step = Erase
	case Erase:
		g.Board.MarkErase()
		g.Board.Erase()
		g.Step = FallStone
	case FallStone:
		if !g.Board.FallStone() {
			g.InitPick()
			g.Step = Move
		} else {
			g.Step = Erase
		}
	}
}

func (g *Game) IsPickCollide(px, py int) bool {
	for i, _ := range g.Pick {
		if a, ok := g.Board.At(px, py-i); ok {
			if *a != nil && (*a).Color != None {
				return true
			}
		}
	}
	return false
}

func (g *Game) FixPick() {
	for i, p := range g.Pick {
		if a, ok := g.Board.At(g.PickX, g.PickY-i); ok {
			if *a == nil || (*a).Color == None {
				*a = p
			} else {
				log.Panic("cell must nil", g.PickX, g.PickY-i, *a)
			}
		} else {
			log.Panic("fix failed", g.PickX, g.PickY-i)
		}
	}
	g.Pick = []*Stone{}
}

func (g *Game) Render(r *ebiten.Image) {
	ebitenutil.DebugPrint(r, "Hello, World!")
	ox := 10
	oy := ScreenHeight - StoneHeight*BoardHeight
	g.Board.Render(r, ox, oy)
	for i, p := range g.Pick {
		g.Board.RenderStone(r, ox, oy, g.PickX, g.PickY-i, p)
	}
}

type Board struct {
	Cell [BoardWidth][BoardHeight]*Stone
}

func (b *Board) Initialize() {
	// Bottom
	for cx := 0; cx < BoardWidth; cx++ {
		if a, ok := b.At(cx, BoardHeight-1); ok {
			*a = NewWall()
		}
	}
	for cy := 0; cy < BoardHeight; cy++ {
		// left
		if a, ok := b.At(0, cy); ok {
			*a = NewWall()
		}
		// right
		if a, ok := b.At(BoardWidth-1, cy); ok {
			*a = NewWall()
		}
	}
}

func (b *Board) MarkErase() {
	// horizontal
	consequentRight := func(cx, cy int) int {
		n := 1
		if c, ok := b.At(cx, cy); ok && *c != nil {
			for x := 1; cx+x < BoardWidth; x++ {
				if c2, ok := b.At(cx+x, cy); ok && *c2 != nil {
					if (*c).Color == (*c2).Color && (*c).Color != Wall {
						n++
						continue
					}
				}
				break
			}
		}
		return n
	}
	for cy := 0; cy < BoardHeight; cy++ {
		for cx := 0; cx < BoardWidth; {
			right := consequentRight(cx, cy)
			if right >= 3 {
				for r := 0; r < right; r++ {
					if c, ok := b.At(cx+r, cy); ok && *c != nil {
						(*c).Erased = true
					}
				}
			}
			cx += right
		}
	}

	// vertical
	consequentDown := func(cx, cy int) int {
		n := 1
		if c, ok := b.At(cx, cy); ok && *c != nil {
			for y := 1; cy+y < BoardHeight; y++ {
				if c2, ok := b.At(cx, cy+y); ok && *c2 != nil {
					if (*c).Color == (*c2).Color && (*c).Color != Wall {
						n++
						continue
					}
				}
				break
			}
		}
		return n
	}
	for cx := 0; cx < BoardWidth; cx++ {
		for cy := 0; cy < BoardHeight; {
			down := consequentDown(cx, cy)
			if down >= 3 {
				for d := 0; d < down; d++ {
					if c, ok := b.At(cx, cy+d); ok && *c != nil {
						(*c).Erased = true
					}
				}
			}
			cy += down
		}
	}

}

func (b *Board) Erase() {
	for cy := 0; cy < BoardHeight; cy++ {
		for cx := 0; cx < BoardWidth; cx++ {
			if c, ok := b.At(cx, cy); ok && (*c) != nil && (*c).Erased {
				*c = nil
			}
		}
	}
}

func (b *Board) FallStone() bool {
	falled := false
	for cx := 0; cx < BoardWidth; cx++ {
		for cy := BoardHeight - 1; cy >= 0; cy-- {
			if c, ok := b.At(cx, cy); ok {
				if c2, ok := b.At(cx, cy-1); ok {
					if (*c) == nil && (*c2) != nil {
						*c, *c2 = *c2, *c
						falled = true
					}
				}
			}
		}
	}
	return falled
}

func (b *Board) At(cx, cy int) (**Stone, bool) {
	if 0 <= cx && cx < BoardWidth {
		if 0 <= cy && cy < BoardHeight {
			return &b.Cell[cx][cy], true
		}
	}
	return nil, false
}

func (b *Board) RenderStone(r *ebiten.Image, ox, oy int, cx, cy int, s *Stone) {
	if s == nil {
		log.Panic("s must not nil")
	}
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Translate(float64(ox), float64(oy))
	opt.GeoM.Translate(float64(cx*StoneWidth), float64(cy*StoneHeight))

	if image, ok := StoneImages[s.Color]; ok {
		err := r.DrawImage(image, opt)
		if err != nil {
			log.Panic(err)
		}
	}

}

func (b *Board) Render(r *ebiten.Image, ox, oy int) {
	for cx := 0; cx < BoardWidth; cx++ {
		for cy := 0; cy < BoardHeight; cy++ {
			opt := &ebiten.DrawImageOptions{}
			opt.GeoM.Translate(float64(ox), float64(oy))
			opt.GeoM.Translate(float64(cx*StoneWidth), float64(cy*StoneHeight))
			// bg
			err := r.DrawImage(StoneImages[None], opt)
			if err != nil {
				log.Panic(err)
			}
			// Stone
			if c, ok := b.At(cx, cy); ok && *c != nil {
				b.RenderStone(r, ox, oy, cx, cy, *c)
			}
		}
	}
}

func NewBoard() *Board {
	return &Board{}
}

func init() {
	StoneImages = make(map[Color]*ebiten.Image)
	sfs, err := fs.New()
	if err != nil {
		log.Panic(err)
	}
	tf, err := sfs.Open("/texture.png")
	if err != nil {
		log.Panic(err)
	}
	defer tf.Close()
	var texture image.Image
	if texture, _, err = image.Decode(tf); err != nil {
		log.Panic(err)
	}
	if Texture, err = ebiten.NewImageFromImage(texture, ebiten.FilterNearest); err != nil {
		log.Panic(err)
	}

	stoneSubImage := func(c Color) *ebiten.Image {
		image := Texture.SubImage(
			image.Rectangle{
				image.Point{StoneWidth * int(c), 0},
				image.Point{StoneWidth*(int(c)+1) - 1, StoneHeight}})
		return image.(*ebiten.Image)
	}
	for _, c := range Colors {
		StoneImages[c] = stoneSubImage(c)
	}

	G = &Game{
		Board: &Board{},
	}
	G.Board.Initialize()
	G.InitPick()
	G.Step = Move
}

func update(screen *ebiten.Image) error {
	G.Update()
	if ebiten.IsDrawingSkipped() {
		return nil
	}
	G.Render(screen)
	return nil
}

func main() {
	if err := ebiten.Run(update, ScreenWidth, ScreenHeight, 2, "Hello, World!"); err != nil {
		log.Fatal(err)
	}
}
