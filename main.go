package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	_ "github.com/neguse/ld44/statik"
	"github.com/rakyll/statik/fs"
)

const Volume = 0.4

type Stone struct {
	Color  Color
	Erased bool
}

func (s *Stone) Colored() bool {
	return s.Color == Red || s.Color == Blue || s.Color == Green || s.Color == Yellow || s.Color == Pink || s.Color == Orange
}

type Color int

const (
	None Color = iota
	Red
	Blue
	Green
	Yellow
	Pink
	Orange
	Dummy3
	Limit
	Wall
	Cursor
	Jammer
)

var Colors []Color = []Color{
	None,
	Red,
	Blue,
	Green,
	Yellow,
	Pink,
	Orange,
	Dummy3,
	Limit,
	Wall,
	Cursor,
	Jammer,
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
	CauseJammer
	GameOver
)

const (
	ScreenWidth  = 200
	ScreenHeight = 300

	BoardWidth  = 8
	BoardHeight = 16

	StoneWidth  = 16
	StoneHeight = 16

	PickMax    = 6
	ReserveNum = PickMax

	JammerTurn = 5

	NumberWidth  = 16
	NumberHeight = 32

	AlphaWidth  = 16
	AlphaHeight = 16

	WaitEraseFrame = 15

	Cross  = 10
	Equal  = 11
	Period = 12
	NumE   = 13
	NumN   = 14
	NumD   = 15
)

var Texture *ebiten.Image
var AudioCtx *audio.Context
var Music *audio.Player
var MusicOff *audio.Player
var StoneImages map[Color]*ebiten.Image
var NumberImages map[int]*ebiten.Image
var AlphaImages map[rune]*ebiten.Image

func PlayMusic(on bool) {
	if on {
		t := MusicOff.Current()
		Music.Seek(t)
		Music.Play()
		MusicOff.Pause()
	} else {
		t := Music.Current()
		MusicOff.Seek(t)
		MusicOff.Play()
		Music.Pause()
	}
}

func PlaySound(s Sound) {
	if s, ok := SoundMap[s]; ok {
		s.SetVolume(Volume)
		s.Rewind()
		s.Play()
	}
}

func (g *Game) Next() *Stone {
	if len(g.Buffer) == 0 {
		level := 3
		if g.Turn > 24 {
			level++
		}
		if g.Turn > 48 {
			level++
		}
		if g.Turn > 72 {
			level++
		}
		colors := []Color{Red, Blue, Green, Yellow, Pink, Orange}[:level]
		rand.Shuffle(len(colors), func(i, j int) {
			colors[i], colors[j] = colors[j], colors[i]
		})
		g.Buffer = colors
	}
	var c Color
	c, g.Buffer = g.Buffer[0], g.Buffer[1:]
	return &Stone{Color: c}
}

type Game struct {
	Board                 *Board
	Buffer                []Color
	Pick                  []*Stone
	PickX, PickY, PickLen int
	Step                  Step
	Wait                  int
	PrevTouchID           int
	MouseEnabled          bool
	DebugString           string

	FirstTouchID        ebiten.TouchID
	FirstTouchPoint     Point
	FirstTouchLastPoint Point
	FirstTouchCursored  bool

	SequentErase  int
	EraseNum      int
	Turn          int
	Score         int
	ScoreEquation string
	HighScore     int
	Ticks         int
}

func NewGame() *Game {
	g := &Game{
		Board:        &Board{},
		MouseEnabled: false,
	}
	g.Initialize()
	return g
}

func (g *Game) Initialize() {
	g.Board.Initialize()
	g.Buffer = nil
	g.Step = Title
	g.Wait = 0
	g.SequentErase = 0
	g.EraseNum = 0
	g.Turn = 0
	g.Score = 0
	g.InitPick()
}

func (g *Game) IsFull() bool {
	for x := 1; x < BoardWidth-1; x++ {
		if g.Board.HeightAt(x) > 1 {
			return false
		}
	}
	return true
}

func (g *Game) HeightAverage() float64 {
	sum := 0.0
	for x := 1; x < BoardWidth-1; x++ {
		sum += float64(g.Board.HeightAt(x))
	}
	return sum / float64(BoardWidth-2)
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
			g.AdjustPick(cx, cy)
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
		}
	}
}

func (g *Game) ReservePick() {
	n := ReserveNum - len(g.Pick)
	for n > 0 {
		g.Pick = append(g.Pick, g.Next())
		n--
	}
}

func NewWall() *Stone {
	return &Stone{
		Color: Wall,
	}
}

func NewJammer() *Stone {
	return &Stone{
		Color: Jammer,
	}
}

func CalcScore(sequent, num int) (int, string) {
	a := 1 << uint(sequent)
	b := num
	score := a * b
	return score, fmt.Sprintf("%dx%d=%d.", a, b, score)
}

func (g *Game) InitPick() {
	g.Pick = nil
	g.ReservePick()
	g.PickX = 3
	g.PickY = PickMax
	g.PickLen = 1
	g.AdjustPick(g.PickX, BoardHeight)
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	} else if v > max {
		return max
	}
	return v
}

func maxInt(v, v2 int) int {
	if v > v2 {
		return v
	}
	return v2
}

func minInt(v, v2 int) int {
	if v > v2 {
		return v2
	}
	return v
}

func (g *Game) AdjustPick(cx, cy int) {
	g.PickX = clampInt(cx, 1, BoardWidth-2)
	height := g.Board.HeightAt(g.PickX)
	g.PickY = PickMax - 1 + minInt(height-PickMax-1, 0)
	g.PickLen = clampInt((g.PickY-maxInt(cy, 0))+1, 0, PickMax)
}

func (g *Game) Update() error {
	g.Ticks++
	switch g.Step {
	case Title:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.MouseEnabled = true
			g.Step = Move
			PlayMusic(true)
		}
		if len(ebiten.TouchIDs()) > 0 {
			g.Step = Move
			PlayMusic(true)
		}
	case Move:
		// move by mouse cursor
		if g.MouseEnabled {
			x, y := ebiten.CursorPosition()
			cx, cy := g.Board.PosToCell(x, y)
			g.AdjustPick(cx, cy)
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				g.FixPick()
			}
		}
		// move by touch
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
				g.Step = CauseJammer
			}
		}
	case FallStone:
		if !g.Board.FallStone() {
			if num := g.Board.MarkErase(); num > 0 {
				g.Wait = WaitEraseFrame
				g.SequentErase++
				g.EraseNum = num
				score, scoreEquation := CalcScore(g.SequentErase, num)
				g.Score += score
				g.HighScore = maxInt(g.HighScore, g.Score)
				g.ScoreEquation = scoreEquation
				if g.SequentErase%4 == 1 {
					PlaySound(S1)
				} else if g.SequentErase%4 == 2 {
					PlaySound(S2)
				} else if g.SequentErase%4 == 3 {
					PlaySound(S3)
				} else if g.SequentErase%4 == 0 {
					PlaySound(S4)
				}
			} else {
				g.Wait = 1
			}
			g.Step = WaitErase
		}
	case CauseJammer:
		g.Turn++
		if g.Turn%JammerTurn == 0 {
			g.CauseJammer()
		}
		g.ReservePick()
		if g.IsFull() {
			g.Step = GameOver
			PlayMusic(false)
		} else {
			g.Step = Move
			g.SequentErase = 0
		}
	case GameOver:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.Initialize()
		}
		if len(ebiten.TouchIDs()) > 0 {
			g.Initialize()
		}
	}
	return nil
}

func (g *Game) CauseJammer() {
	num := (g.Turn/JammerTurn+2)%3 + 1
	if g.Turn > 50 {
		num++
	}
	for i := 0; i < num; i++ {
		x := rand.Intn(BoardWidth-2) + 1
		y := g.Board.HeightAt(x) - 1
		if y > 1 {
			if c, ok := g.Board.At(x, y); ok {
				*c = NewJammer()
			} else {
			}
		} else {
			continue
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
	if g.PickLen > 0 {
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
		g.PickY -= g.PickLen
		g.Pick = g.Pick[g.PickLen:]
		g.PickLen = 1
		g.Step = FallStone
	}
}

func (g *Game) Draw(r *ebiten.Image) {
	r.Fill(color.Gray{Y: 0x80})
	/*
		var input string
		if g.MouseEnabled {
			input = "Click"
		} else {
			input = "Tap twice"
		}
		if g.Step == GameOver {
			ebitenutil.DebugPrint(r, "Game is over")
		} else {
			ebitenutil.DebugPrint(r, "  "+input+" to cut! match 3!"+"\n"+g.DebugString)
		}
	*/
	g.DebugString = ""
	avg := g.HeightAverage()
	noise := math.Max((8.0-avg)*0.2, 0.0)
	g.Board.Render(r, noise, g.Wait)
	if g.Step != Title {
		for i, p := range g.Pick {
			cx, cy := g.PickX, g.PickY-i
			if cy >= 0 {
				g.Board.RenderStone(r, cx, cy, p, noise, g.Wait)
				if i+1 == g.PickLen && g.Step == Move {
					g.Board.RenderCursor(r, cx, cy)
				}
			}
		}
		if g.SequentErase > 0 {
			f := (float64(g.Wait) / WaitEraseFrame)
			dx := f * f * f * NumberWidth
			RenderNumber(r, g.SequentErase, BoardWidth*StoneWidth/2+NumberWidth+int(dx), 0, false)
			RenderEquation(r, g.ScoreEquation, ScreenWidth, ScreenHeight-32, true)
		} else {
			RenderNumber(r, g.Score, ScreenWidth, ScreenHeight-32, true)
		}
	}
	if g.Step == GameOver {
		RenderEnd(r, BoardWidth*StoneWidth/2-NumberWidth, StoneHeight*3, g.Ticks)
	}
	if g.Step == Title {
		// ebitenutil.DebugPrint(r, "\n  cut'n'align\n  LD44 game by @neguse\n 2019 end of heisei generation\n\n\n\n  click to start\n\n\n\n\n\n\n  Very thanks to \n    @hajimehoshi\n    and my brother.")
		ebitenutil.DebugPrintAt(r, "Very thanks to\n@hajimehoshi\nand my brother.", 32, ScreenHeight-60)
		RenderNumber(r, g.HighScore, ScreenWidth, ScreenHeight-32, true)
		RenderAlpha(r, "cutn", StoneWidth*1.5, StoneHeight*3)
		RenderAlpha(r, "align", StoneWidth*2.5, StoneHeight*4)
		RenderAlpha(r, "click", StoneWidth*1.5, StoneHeight*6)
		RenderAlpha(r, "to", StoneWidth*3.5, StoneHeight*7)
		RenderAlpha(r, "cut", StoneWidth*2.5, StoneHeight*8)
		// RenderAlpha(r, "@@@@@@", StoneWidth*1.5, StoneHeight*12+1)
		RenderAlpha(r, "neguse", StoneWidth*1.5, StoneHeight*13+1)
	}

}

type Board struct {
	Cell             [BoardWidth][BoardHeight]*Stone
	OriginX, OriginY int
}

func (b *Board) Initialize() {
	for cx := 0; cx < BoardWidth; cx++ {
		for cy := 0; cy < BoardHeight; cy++ {
			if c, ok := b.At(cx, cy); ok {
				*c = nil
			}
		}
	}
	// Bottom
	for cx := 0; cx < BoardWidth; cx++ {
		if c, ok := b.At(cx, BoardHeight-1); ok {
			*c = NewWall()
		}
	}
	for cy := 0; cy < BoardHeight; cy++ {
		// left
		if c, ok := b.At(0, cy); ok {
			*c = NewWall()
		}
		// right
		if c, ok := b.At(BoardWidth-1, cy); ok {
			*c = NewWall()
		}
	}
	b.OriginX = 10
	b.OriginY = -10 + ScreenHeight - StoneHeight*BoardHeight
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

func (b *Board) MarkEraseAt(cx, cy int) bool {
	if c, ok := b.At(cx, cy); ok && *c != nil {
		(*c).Erased = true

		// erase jammer next of erased
		dp := []Point{
			Point{-1, 0},
			Point{1, 0},
			Point{0, -1},
			Point{0, 1},
		}
		for _, d := range dp {
			ncx, ncy := cx+d.x, cy+d.y
			if c, ok := b.At(ncx, ncy); ok && *c != nil && (*c).Color == Jammer {
				(*c).Erased = true
			}
		}

		return true
	}
	return false
}

func (b *Board) MarkErase() int {
	num := 0
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
						if (*c).Color == (*c2).Color && (*c).Colored() {
							continue
						}
					}
				}
			}
			n := i - sequent
			if n >= 3 {
				for _, cp := range line[sequent:i] {
					if b.MarkEraseAt(cp.x, cp.y) {
						num++
					}
				}
			}
			sequent = i
		}
	}
	return num
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

func (b *Board) RenderStone(r *ebiten.Image, cx, cy int, s *Stone, noise float64, wait int) {
	if s == nil {
		log.Panic("s must not nil")
	}
	opt := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	// sugoi nazo no erasing animation
	if s.Erased {
		opt.GeoM.Translate(-float64(StoneWidth*0.5)+3.0, -float64(StoneHeight)*0.5)
		r := float64(wait)
		opt.GeoM.Rotate(r)
		s := float64(wait) / float64(WaitEraseFrame)
		opt.GeoM.Scale(s*s*s, s*s*s)
		opt.GeoM.Translate(float64(StoneWidth*0.5), float64(StoneHeight)*0.5)
	}
	opt.GeoM.Translate(float64(b.OriginX)+(rand.Float64()-0.5)*noise, float64(b.OriginY)+(rand.Float64()-0.5)*noise)
	opt.GeoM.Translate(float64(cx*StoneWidth), float64(cy*StoneHeight))

	if image, ok := StoneImages[s.Color]; ok {
		r.DrawImage(image, opt)
	}
}

func (b *Board) PosToCell(x, y int) (cx, cy int) {
	return (x - b.OriginX) / StoneWidth, (y - b.OriginY) / StoneHeight
}

func (b *Board) HeightAt(x int) int {
	for y := 0; y < BoardHeight; y++ {
		if c, ok := b.At(x, y); !ok || (*c) == nil {
			continue
		}
		return y
	}
	return BoardHeight
}

func (b *Board) RenderCursor(r *ebiten.Image, cx, cy int) {
	opt := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	opt.GeoM.Translate(float64(b.OriginX), float64(b.OriginY))
	opt.GeoM.Translate(float64(cx*StoneWidth), float64(cy*StoneHeight))

	if image, ok := StoneImages[Cursor]; ok {
		r.DrawImage(image, opt)
	}
}

func (b *Board) Render(r *ebiten.Image, noise float64, wait int) {
	for cx := 0; cx < BoardWidth; cx++ {
		for cy := 0; cy < BoardHeight; cy++ {
			opt := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
			opt.GeoM.Translate(float64(b.OriginX)+(rand.Float64()-0.5)*noise, float64(b.OriginY)+(rand.Float64()-0.5)*noise)
			opt.GeoM.Translate(float64(cx*StoneWidth), float64(cy*StoneHeight))
			// bg
			if cy == 0 {
				r.DrawImage(StoneImages[Limit], opt)
			} else {
				r.DrawImage(StoneImages[None], opt)
			}

			// Stone
			if c, ok := b.At(cx, cy); ok && *c != nil {
				b.RenderStone(r, cx, cy, *c, noise, wait)
			}
		}
	}
}

// x, y is right bottom
func RenderEquation(r *ebiten.Image, equation string, x, y int, rot bool) {
	ctoi := func(ch rune) int {
		switch ch {
		case 'x':
			return Cross
		case '=':
			return Equal
		case '.':
			return Period
		default:
			return int(ch) - int('0')
		}
	}
	for i, c := range equation {
		opt := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		if rot {
			opt.GeoM.Rotate(math.Pi / 2)
		}
		opt.GeoM.Translate(float64(x-NumberWidth), float64(y+(-len(equation)+i+1)*NumberWidth))
		r.DrawImage(NumberImages[ctoi(c)], opt)
	}
}

// x, y is right bottom
func RenderAlpha(r *ebiten.Image, str string, x, y int) {
	for i, c := range str {
		opt := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		opt.GeoM.Translate(float64(x+AlphaWidth*i), float64(y))
		r.DrawImage(AlphaImages[c], opt)
	}
}

// perhaps x, y is right bottom
func RenderNumber(r *ebiten.Image, n int, x, y int, rot bool) {
	opt := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	if rot {
		opt.GeoM.Rotate(math.Pi / 2)
	}
	opt.GeoM.Translate(float64(x-NumberWidth), float64(y))
	r.DrawImage(NumberImages[n%10], opt)
	if n >= 10 {
		RenderNumber(r, n/10, x, y-NumberWidth, rot)
	}
}

func RenderEnd(r *ebiten.Image, x, y int, ticks int) {
	for i, n := range []int{NumE, NumN, NumD} {
		ny := (math.Cos((float64(ticks)+float64(i))*0.1) + 1.0) * float64(BoardHeight*StoneHeight) * 0.25
		opt := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		opt.GeoM.Translate(float64(x+NumberWidth*i), float64(y)+ny)
		r.DrawImage(NumberImages[n], opt)
	}
}

func NewBoard() *Board {
	return &Board{}
}

func init() {
	StoneImages = make(map[Color]*ebiten.Image)
	NumberImages = make(map[int]*ebiten.Image)
	AlphaImages = make(map[rune]*ebiten.Image)
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
	if texture, err = png.Decode(tf); err != nil {
		log.Panic(err)
	}
	Texture = ebiten.NewImageFromImage(texture)

	stoneSubImage := func(i int) *ebiten.Image {
		x := i % 8
		y := i / 8
		image := Texture.SubImage(
			image.Rectangle{
				image.Point{StoneWidth * x, StoneHeight * y},
				image.Point{StoneWidth * (x + 1), StoneHeight * (y + 1)}})
		return image.(*ebiten.Image)
	}
	for _, c := range Colors {
		StoneImages[c] = stoneSubImage(int(c))
	}
	numberSubImage := func(i int) *ebiten.Image {
		x := i % 8
		y := i / 8
		image := Texture.SubImage(
			image.Rectangle{
				image.Point{NumberWidth * x, 32 + NumberHeight*y},
				image.Point{NumberWidth * (x + 1), 32 + NumberHeight*(y+1)}})
		return image.(*ebiten.Image)
	}
	for i := 0; i < 16; i++ {
		NumberImages[i] = numberSubImage(i)
	}
	alphaSubImage := func(i int) *ebiten.Image {
		x := i % 8
		y := i / 8
		image := Texture.SubImage(
			image.Rectangle{
				image.Point{AlphaWidth * x, 96 + AlphaHeight*y},
				image.Point{AlphaWidth * (x + 1), 96 + AlphaHeight*(y+1)}})
		return image.(*ebiten.Image)
	}
	for i, c := range []rune{'c', 'u', 't', 'n', 'a', 'l', 'i', 'g', 'k', 'o', '@', 'e', 's'} {
		AlphaImages[c] = alphaSubImage(i)
	}

	AudioCtx = audio.NewContext(44100)

	{
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
	}
	{
		mf, err := sfs.Open("/bgm_off.ogg")
		if err != nil {
			log.Panic(err)
		}
		defer mf.Close()
		s, err := vorbis.Decode(AudioCtx, mf)
		if err != nil {
			log.Panic(err)
		}
		MusicOff, err = audio.NewPlayer(AudioCtx, audio.NewInfiniteLoop(s, s.Length()))
		if err != nil {
			log.Panic(err)
		}
		MusicOff.SetVolume(Volume)
	}

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
		p := audio.NewPlayerFromBytes(AudioCtx, vdata)
		SoundMap[s] = p
	}

	PlayMusic(false)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetMaxTPS(30)
	ebiten.SetWindowTitle("cut'n'align")
	g := NewGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
