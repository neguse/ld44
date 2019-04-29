package main

import (
	"image"
	"io/ioutil"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/vorbis"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"

	_ "github.com/neguse/ld44/statik"
	"github.com/rakyll/statik/fs"
)

const Volume = 0.4

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
	Cursor
)

var Colors []Color = []Color{
	None,
	Red,
	Blue,
	Green,
	Yellow,
	Wall,
	Cursor,
}

type Sound int

const (
	S1 Sound = iota
	S2
	S3
	S4
)

var SoundNameMap map[string]Sound = map[string]Sound{
	"/S1.ogg": S1,
	"/S2.ogg": S2,
	"/S3.ogg": S3,
	"/S4.ogg": S4,
}

var SoundMap map[Sound]*audio.Player = map[Sound]*audio.Player{}

type Step int

const (
	Title Step = iota
	Move
	FallStone
	WaitErase
)

const (
	ScreenWidth  = 200
	ScreenHeight = 300

	BoardWidth  = 8
	BoardHeight = 16

	StoneWidth  = 16
	StoneHeight = 16

	PickMin    = 2
	PickMax    = 6
	ReserveNum = PickMax
)

var Texture *ebiten.Image
var AudioCtx *audio.Context
var Music *audio.Player
var StoneImages map[Color]*ebiten.Image
var G *Game

func PlaySound(s Sound) {
	if s, ok := SoundMap[s]; ok {
		s.SetVolume(Volume)
		s.Rewind()
		s.Play()
	}
}

type StoneGenerator struct {
}

func (g *StoneGenerator) Next() *Stone {
	// TODO correct rate
	r := rand.Intn(4)
	return &Stone{
		Color: []Color{Red, Blue, Green, Yellow}[r],
	}
}

type Game struct {
	Board                 *Board
	Gen                   *StoneGenerator
	Pick                  []*Stone
	PickX, PickY, PickLen int
	Step                  Step
	Wait                  int
	PrevTouchID           int
	MouseEnabled          bool
	DebugString           string

	FirstTouchID        int
	FirstTouchPoint     Point
	FirstTouchLastPoint Point
	FirstTouchCursored  bool

	SequentErase int
}

func (g *Game) UpdateTouch() {
	for _, tid := range ebiten.TouchIDs() {
		if g.FirstTouchID == 0 {
			g.FirstTouchID = tid
			x, y := ebiten.TouchPosition(tid)
			g.FirstTouchPoint = Point{x, y}
			cx, cy := g.Board.PosToCell(x, y)
			g.FirstTouchCursored = cx == g.PickX && cy == (g.PickY-g.PickLen)+1
		}
		if tid == g.FirstTouchID {
			x, y := ebiten.TouchPosition(tid)
			cx, cy := g.Board.PosToCell(x, y)
			g.PickX = clampInt(cx, 1, BoardWidth-2)
			g.PickLen = clampInt((g.PickY-cy)+1, 1, PickMax)
			g.FirstTouchLastPoint = Point{x, y}
		}
	}
	if inpututil.IsTouchJustReleased(g.FirstTouchID) {
		g.FirstTouchID = 0
		x, y := g.FirstTouchLastPoint.x, g.FirstTouchLastPoint.y
		pcx, pcy := g.Board.PosToCell(g.FirstTouchPoint.x, g.FirstTouchPoint.y)
		cx, cy := g.Board.PosToCell(x, y)
		cursored := cx == g.PickX && cy == (g.PickY-g.PickLen)+1
		if cursored && g.FirstTouchCursored && pcx == cx && pcy == cy {
			g.FixPick()
			g.PickLen = 1
			g.Step = FallStone
		}
	}
}

func (g *Game) ReservePick() {
	n := ReserveNum - len(g.Pick)
	for n > 0 {
		g.Pick = append(g.Pick, g.Gen.Next())
		n--
	}
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
	g.ReservePick()
	g.PickX = 3
	g.PickY = PickMax
	g.PickLen = 1
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	} else if v > max {
		return max
	}
	return v
}

func (g *Game) Update() {
	switch g.Step {
	case Title:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.MouseEnabled = true
			g.Step = Move
		}
		if len(ebiten.TouchIDs()) > 0 {
			g.Step = Move
		}
	case Move:

		// move by mouse cursor
		if g.MouseEnabled {
			x, y := ebiten.CursorPosition()
			cx, cy := g.Board.PosToCell(x, y)
			g.PickX = clampInt(cx, 1, BoardWidth-2)
			g.PickLen = clampInt((g.PickY-cy)+1, 1, PickMax)
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				g.FixPick()
				g.PickLen = 1
				g.Step = FallStone
			}
		}

		g.UpdateTouch()

		/*
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
				if g.PickLen < PickMax-1 {
					g.PickLen++
				}
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
				if g.PickLen > 1 {
					g.PickLen--
				}
			}
			if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				g.FixPick()
				g.PickLen = 1
				g.Step = FallStone
			}
		*/
	case WaitErase:
		g.Wait--
		if g.Wait <= 0 {
			if g.Board.Erase() {
				g.Step = FallStone
			} else {
				g.ReservePick()
				g.Step = Move
				g.SequentErase = 0
			}
		}
	case FallStone:
		if !g.Board.FallStone() {
			if g.Board.MarkErase() {
				g.Wait = 10
				g.SequentErase++
				if g.SequentErase == 1 {
					PlaySound(S1)
				} else if g.SequentErase == 2 {
					PlaySound(S2)
				} else if g.SequentErase == 3 {
					PlaySound(S3)
				} else if g.SequentErase >= 4 {
					PlaySound(S4)
				}
			} else {
				g.Wait = 1
			}
			g.Step = WaitErase
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
	for i, p := range g.Pick[:g.PickLen] {
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
	g.Pick = g.Pick[g.PickLen:]
}

func (g *Game) Render(r *ebiten.Image) {
	if g.Step == Title {
		ebitenutil.DebugPrint(r, "  LD44 game by neguse\n  click to start\n  TODO: choose right title.")
	} else {
		var input string
		if g.MouseEnabled {
			input = "Click"
		} else {
			input = "Tap twice"
		}
		ebitenutil.DebugPrint(r, "  "+input+" to cut! match 3!"+"\n"+g.DebugString)
		g.DebugString = ""
		g.Board.Render(r)
		for i, p := range g.Pick {
			cx, cy := g.PickX, g.PickY-i
			g.Board.RenderStone(r, cx, cy, p)
			if i+1 == g.PickLen {
				g.Board.RenderCursor(r, cx, cy)
			}
		}
	}
}

type Board struct {
	Cell             [BoardWidth][BoardHeight]*Stone
	OriginX, OriginY int
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

	b.OriginX = 10
	b.OriginY = ScreenHeight - StoneHeight*BoardHeight
}

type Point struct {
	x, y int
}

func HorizontalLines() [][]Point {
	var lines [][]Point
	for cy := 0; cy < BoardHeight; cy++ {
		var line []Point
		for cx := 0; cx < BoardWidth; cx++ {
			line = append(line, Point{cx, cy})
		}
		lines = append(lines, line)
	}
	return lines
}

func VerticalLines() [][]Point {
	var lines [][]Point
	for cx := 0; cx < BoardWidth; cx++ {
		var line []Point
		for cy := 0; cy < BoardHeight; cy++ {
			line = append(line, Point{cx, cy})
		}
		lines = append(lines, line)
	}
	return lines
}

func RightDownLines() [][]Point {
	rightDownLine := func(x, y int) []Point {
		var line []Point
		for i := 0; ; i++ {
			cx, cy := x+i, y+i
			if cx >= BoardWidth || cy >= BoardHeight {
				break
			}
			line = append(line, Point{cx, cy})
		}
		return line
	}
	var lines [][]Point
	for y := 0; y < BoardHeight; y++ {
		lines = append(lines, rightDownLine(0, y))
	}
	for x := 1; x < BoardWidth; x++ {
		lines = append(lines, rightDownLine(x, 0))
	}
	return lines
}

func RightUpLines() [][]Point {
	rightUpLine := func(x, y int) []Point {
		var line []Point
		for i := 0; ; i++ {
			cx, cy := x+i, y-i
			if cx >= BoardWidth || cy < 0 {
				break
			}
			line = append(line, Point{cx, cy})
		}
		return line
	}
	var lines [][]Point
	for y := 0; y < BoardHeight; y++ {
		lines = append(lines, rightUpLine(0, y))
	}
	for x := 1; x < BoardWidth; x++ {
		lines = append(lines, rightUpLine(x, BoardHeight-1))
	}
	return lines
}

func (b *Board) MarkErase() bool {
	erased := false
	var lines [][]Point

	lines = append(lines, HorizontalLines()...)
	lines = append(lines, VerticalLines()...)
	lines = append(lines, RightDownLines()...)
	lines = append(lines, RightUpLines()...)

	for _, line := range lines {
		sequent := 0
		for i := 1; i <= len(line); i++ {
			p := line[sequent]
			if i < len(line) {
				p2 := line[i]
				if c, ok := b.At(p.x, p.y); ok && *c != nil {
					if c2, ok := b.At(p2.x, p2.y); ok && *c2 != nil {
						if (*c).Color == (*c2).Color && (*c).Color != Wall {
							continue
						}
					}
				}
			}
			n := i - sequent
			if n >= 3 {
				for _, cp := range line[sequent:i] {
					if c, ok := b.At(cp.x, cp.y); ok && *c != nil {
						(*c).Erased = true
						erased = true
					}
				}
			}
			sequent = i
		}
	}
	return erased
}

func (b *Board) Erase() bool {
	erased := false
	for cy := 0; cy < BoardHeight; cy++ {
		for cx := 0; cx < BoardWidth; cx++ {
			if c, ok := b.At(cx, cy); ok && (*c) != nil && (*c).Erased {
				*c = nil
				erased = true
			}
		}
	}
	return erased
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

func (b *Board) RenderStone(r *ebiten.Image, cx, cy int, s *Stone) {
	if s == nil {
		log.Panic("s must not nil")
	}
	opt := &ebiten.DrawImageOptions{}
	// sugoi nazo no erasing animation
	if s.Erased {
		s := float64(G.Wait) / 3.0
		opt.GeoM.Scale(s, s)
	}
	opt.GeoM.Translate(float64(b.OriginX), float64(b.OriginY))
	opt.GeoM.Translate(float64(cx*StoneWidth), float64(cy*StoneHeight))

	if image, ok := StoneImages[s.Color]; ok {
		err := r.DrawImage(image, opt)
		if err != nil {
			log.Panic(err)
		}
	}
}

func (b *Board) PosToCell(x, y int) (cx, cy int) {
	return (x - b.OriginX) / StoneWidth, (y - b.OriginY) / StoneHeight
}

func (b *Board) RenderCursor(r *ebiten.Image, cx, cy int) {
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Translate(float64(b.OriginX), float64(b.OriginY))
	opt.GeoM.Translate(float64(cx*StoneWidth), float64(cy*StoneHeight))

	if image, ok := StoneImages[Cursor]; ok {
		err := r.DrawImage(image, opt)
		if err != nil {
			log.Panic(err)
		}
	}
}

func (b *Board) Render(r *ebiten.Image) {
	for cx := 0; cx < BoardWidth; cx++ {
		for cy := 0; cy < BoardHeight; cy++ {
			opt := &ebiten.DrawImageOptions{}
			opt.GeoM.Translate(float64(b.OriginX), float64(b.OriginY))
			opt.GeoM.Translate(float64(cx*StoneWidth), float64(cy*StoneHeight))
			// bg
			err := r.DrawImage(StoneImages[None], opt)
			if err != nil {
				log.Panic(err)
			}
			// Stone
			if c, ok := b.At(cx, cy); ok && *c != nil {
				b.RenderStone(r, cx, cy, *c)
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
				image.Point{StoneWidth * (int(c) + 1), StoneHeight}})
		return image.(*ebiten.Image)
	}
	for _, c := range Colors {
		StoneImages[c] = stoneSubImage(c)
	}

	AudioCtx, err = audio.NewContext(44100)
	if err != nil {
		log.Panic(err)
	}

	mf, err := sfs.Open("/bgm.ogg")
	if err != nil {
		log.Panic(err)
	}
	defer mf.Close()
	s, err := vorbis.Decode(AudioCtx, mf)
	if err != nil {
		log.Panic(err)
	}
	Music, err = audio.NewPlayer(AudioCtx, audio.NewInfiniteLoop(s, s.Length()))
	if err != nil {
		log.Panic(err)
	}
	Music.SetVolume(Volume)
	Music.Play()

	for sname, s := range SoundNameMap {
		sf, err := sfs.Open(sname)
		if err != nil {
			log.Panic(err)
		}
		defer sf.Close()
		v, err := vorbis.Decode(AudioCtx, sf)
		if err != nil {
			log.Panic(err)
		}
		vdata, err := ioutil.ReadAll(v)
		if err != nil {
			log.Panic(err)
		}
		p, err := audio.NewPlayerFromBytes(AudioCtx, vdata)
		if err != nil {
			log.Panic(err)
		}
		SoundMap[s] = p
	}

	G = &Game{
		Board:        &Board{},
		MouseEnabled: false,
	}
	G.Board.Initialize()
	G.InitPick()
	G.Step = Title
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
	ebiten.SetMaxTPS(30)
	if err := ebiten.Run(update, ScreenWidth, ScreenHeight, 2, "Hello, World!"); err != nil {
		log.Fatal(err)
	}
}
