package main

import (
	"image"
	"log"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"

	_ "github.com/neguse/ld44/statik"
	"github.com/rakyll/statik/fs"
)

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
	BoardHeight = 8

	StoneWidth  = 16
	StoneHeight = 16
)

var Texture *ebiten.Image
var StoneImages map[Color]*ebiten.Image
var G *Game

type Game struct {
	Board        *Board
	Pick         []Color
	PickX, PickY int
	Step         Step
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
			g.Pick = append(g.Pick, Red)
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
		g.Step = FallStone
	case FallStone:
		g.Pick = []Color{Red, Red}
		g.PickX = 3
		g.PickY = -1
		g.Step = Move
	}
}

func (g *Game) IsPickCollide(px, py int) bool {
	for i, _ := range g.Pick {
		if a, ok := g.Board.At(px, py+i); ok {
			if *a != None {
				return true
			}
		}
	}
	return false
}

func (g *Game) FixPick() {
	for i, p := range g.Pick {
		if a, ok := g.Board.At(g.PickX, g.PickY+i); ok {
			if *a == None {
				*a = p
			}
		}
	}
	g.Pick = []Color{}

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
	Cell [BoardWidth][BoardHeight]Color
}

func (b *Board) Initialize() {
	// Bottom
	for cx := 0; cx < BoardWidth; cx++ {
		if a, ok := b.At(cx, BoardHeight-1); ok {
			*a = Wall
		}
	}
	for cy := 0; cy < BoardWidth; cy++ {
		// left
		if a, ok := b.At(0, cy); ok {
			*a = Wall
		}
		// right
		if a, ok := b.At(BoardWidth-1, cy); ok {
			*a = Wall
		}
	}
}

func (b *Board) MarkErase() {

}

func (b *Board) At(cx, cy int) (*Color, bool) {
	if 0 <= cx && cx < BoardWidth {
		if 0 <= cy && cy < BoardHeight {
			return &b.Cell[cx][cy], true
		}
	}
	return nil, false
}

func (b *Board) RenderStone(r *ebiten.Image, ox, oy int, cx, cy int, c Color) {
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Translate(float64(ox), float64(oy))
	opt.GeoM.Translate(float64(cx*StoneWidth), float64(cy*StoneHeight))

	if image, ok := StoneImages[c]; ok {
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
			if c, ok := b.At(cx, cy); ok {
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
	G.PickX = 3
	G.PickY = -1
	G.Pick = []Color{Red, Red}
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
