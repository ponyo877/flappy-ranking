// Copyright 2018 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	raudio "github.com/hajimehoshi/ebiten/v2/examples/resources/audio"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	resources "github.com/hajimehoshi/ebiten/v2/examples/resources/images/flappy"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/ponyo877/flappy-standings/config"
)

func floorDiv(x, y int) int {
	d := x / y
	if d*y == x || x >= 0 {
		return d
	}
	return d - 1
}

func floorMod(x, y int) int {
	return x - floorDiv(x, y)*y
}

// const (
// 	screenWidth      = 640
// 	screenHeight     = 480
// 	tileSize         = 32
// 	titleFontSize    = fontSize * 1.5
// 	fontSize         = 24
// 	smallFontSize    = fontSize / 2
// 	pipeWidth        = tileSize * 2
// 	pipeStartOffsetX = 8
// 	pipeIntervalX    = 8
// 	pipeGapY         = 5
// 	initialX16       = 0
// 	initialY16       = 100 * 16
// 	initialCameraX   = -240
// 	initialCameraY   = 0
// 	unit             = 16
// 	vyLimit          = 96
// 	deltaX16         = 32
// 	deltaVY16        = 4
// 	deltaCameraX     = 2
// )

var (
	gopherImage      *ebiten.Image
	tilesImage       *ebiten.Image
	arcadeFaceSource *text.GoTextFaceSource
)

func init() {
	img, _, err := image.Decode(bytes.NewReader(resources.Gopher_png))
	if err != nil {
		log.Fatal(err)
	}
	gopherImage = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(resources.Tiles_png))
	if err != nil {
		log.Fatal(err)
	}
	tilesImage = ebiten.NewImageFromImage(img)
}

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf))
	if err != nil {
		log.Fatal(err)
	}
	arcadeFaceSource = s
}

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
)

type Game struct {
	mode Mode

	// The gopher's position
	x16  int
	y16  int
	vy16 int

	// Camera
	cameraX int
	cameraY int

	// Pipes
	pipeTileYs []int

	gameoverCount int

	touchIDs   []ebiten.TouchID
	gamepadIDs []ebiten.GamepadID

	audioContext *audio.Context
	jumpPlayer   *audio.Player
	hitPlayer    *audio.Player

	jumpHistory []int
}

func NewGame() ebiten.Game {
	g := &Game{}
	g.init()
	return g
}

func (g *Game) init() {
	g.x16 = config.InitialX16
	g.y16 = config.InitialY16
	g.cameraX = config.InitialCameraX
	g.cameraY = config.InitialCameraY
	g.pipeTileYs = make([]int, 256)
	randomKey := "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"
	seed := [32]byte([]byte(randomKey))
	r := rand.New(rand.NewChaCha8(seed))
	for i := range g.pipeTileYs {
		g.pipeTileYs[i] = r.IntN(6) + 2
	}

	if g.audioContext == nil {
		g.audioContext = audio.NewContext(48000)
	}

	jumpD, err := vorbis.DecodeF32(bytes.NewReader(raudio.Jump_ogg))
	if err != nil {
		log.Fatal(err)
	}
	g.jumpPlayer, err = g.audioContext.NewPlayerF32(jumpD)
	if err != nil {
		log.Fatal(err)
	}

	jabD, err := wav.DecodeF32(bytes.NewReader(raudio.Jab_wav))
	if err != nil {
		log.Fatal(err)
	}
	g.hitPlayer, err = g.audioContext.NewPlayerF32(jabD)
	if err != nil {
		log.Fatal(err)
	}
	g.jumpHistory = []int{}
}

func (g *Game) isKeyJustPressed() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return true
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return true
	}
	g.touchIDs = inpututil.AppendJustPressedTouchIDs(g.touchIDs[:0])
	if len(g.touchIDs) > 0 {
		return true
	}
	g.gamepadIDs = ebiten.AppendGamepadIDs(g.gamepadIDs[:0])
	for _, g := range g.gamepadIDs {
		if ebiten.IsStandardGamepadLayoutAvailable(g) {
			if inpututil.IsStandardGamepadButtonJustPressed(g, ebiten.StandardGamepadButtonRightBottom) {
				return true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(g, ebiten.StandardGamepadButtonRightRight) {
				return true
			}
		} else {
			// The button 0/1 might not be A/B buttons.
			if inpututil.IsGamepadButtonJustPressed(g, ebiten.GamepadButton0) {
				return true
			}
			if inpututil.IsGamepadButtonJustPressed(g, ebiten.GamepadButton1) {
				return true
			}
		}
	}
	return false
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

func (g *Game) Update() error {
	switch g.mode {
	case ModeTitle:
		if g.isKeyJustPressed() {
			g.mode = ModeGame
		}
	case ModeGame:
		g.x16 += config.DeltaX16
		g.cameraX += config.DeltaCameraX
		if g.isKeyJustPressed() {
			g.jumpHistory = append(g.jumpHistory, g.x16)
			g.vy16 = -config.VyLimit
			if err := g.jumpPlayer.Rewind(); err != nil {
				return err
			}
			g.jumpPlayer.Play()
		}
		g.y16 += g.vy16

		// Gravity
		g.vy16 += config.DeltaVy16
		if g.vy16 > config.VyLimit {
			g.vy16 = config.VyLimit
		}

		if g.hit() {
			g.jumpHistory = []int{}
			if err := g.hitPlayer.Rewind(); err != nil {
				return err
			}
			g.hitPlayer.Play()
			g.mode = ModeGameOver
			g.gameoverCount = 30
		}
	case ModeGameOver:
		if g.gameoverCount > 0 {
			g.gameoverCount--
		}
		if g.gameoverCount == 0 && g.isKeyJustPressed() {
			g.init()
			g.mode = ModeTitle
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})
	g.drawTiles(screen)
	if g.mode != ModeTitle {
		g.drawGopher(screen)
	}

	var titleTexts string
	var texts string
	switch g.mode {
	case ModeTitle:
		titleTexts = "FLAPPY GOPHER\nWITH STANDIGNS"
		texts = "\n\n\n\n\n\nPRESS SPACE KEY\n\nOR A/B BUTTON\n\nOR TOUCH SCREEN"
	case ModeGameOver:
		texts = "\nGAME OVER!"
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(config.ScreenWidth/2, 3*config.TitleFontSize)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = config.TitleFontSize
	op.PrimaryAlign = text.AlignCenter
	text.Draw(screen, titleTexts, &text.GoTextFace{
		Source: arcadeFaceSource,
		Size:   config.TitleFontSize,
	}, op)

	op = &text.DrawOptions{}
	op.GeoM.Translate(config.ScreenWidth/2, 3*config.TitleFontSize)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = config.FontSize
	op.PrimaryAlign = text.AlignCenter
	text.Draw(screen, texts, &text.GoTextFace{
		Source: arcadeFaceSource,
		Size:   config.FontSize,
	}, op)

	if g.mode == ModeTitle {
		const msg = "Go Gopher by Renee French is\nlicenced under CC BY 3.0."

		op := &text.DrawOptions{}
		op.GeoM.Translate(config.ScreenWidth/2, config.ScreenHeight-config.SmallFontSize/2)
		op.ColorScale.ScaleWithColor(color.White)
		op.LineSpacing = config.SmallFontSize
		op.PrimaryAlign = text.AlignCenter
		op.SecondaryAlign = text.AlignEnd
		text.Draw(screen, msg, &text.GoTextFace{
			Source: arcadeFaceSource,
			Size:   config.SmallFontSize,
		}, op)
	}

	op = &text.DrawOptions{}
	op.GeoM.Translate(config.ScreenWidth, 0)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = config.FontSize
	op.PrimaryAlign = text.AlignEnd
	text.Draw(screen, fmt.Sprintf("%04d", g.score()), &text.GoTextFace{
		Source: arcadeFaceSource,
		Size:   config.FontSize,
	}, op)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func (g *Game) pipeAt(tileX int) (tileY int, ok bool) {
	if (tileX - config.PipeStartOffsetX) <= 0 {
		return 0, false
	}
	if floorMod(tileX-config.PipeStartOffsetX, config.PipeIntervalX) != 0 {
		return 0, false
	}
	idx := floorDiv(tileX-config.PipeStartOffsetX, config.PipeIntervalX)
	return g.pipeTileYs[idx%len(g.pipeTileYs)], true
}

func (g *Game) score() int {
	x := floorDiv(g.x16, config.Unit) / config.TileSize
	if (x - config.PipeStartOffsetX) <= 0 {
		return 0
	}
	return floorDiv(x-config.PipeStartOffsetX, config.PipeIntervalX)
}

func (g *Game) hit() bool {
	if g.mode != ModeGame {
		return false
	}
	const (
		gopherWidth  = 30
		gopherHeight = 60
	)
	w, h := gopherImage.Bounds().Dx(), gopherImage.Bounds().Dy()
	x0 := floorDiv(g.x16, config.Unit) + (w-gopherWidth)/2
	y0 := floorDiv(g.y16, config.Unit) + (h-gopherHeight)/2
	x1 := x0 + gopherWidth
	y1 := y0 + gopherHeight
	if y0 < -config.TileSize*4 {
		return true
	}
	if y1 >= config.ScreenHeight-config.TileSize {
		return true
	}
	xMin := floorDiv(x0-config.PipeWidth, config.TileSize)
	xMax := floorDiv(x0+gopherWidth, config.TileSize)
	for x := xMin; x <= xMax; x++ {
		y, ok := g.pipeAt(x)
		if !ok {
			continue
		}
		if x0 >= x*config.TileSize+config.PipeWidth {
			continue
		}
		if x1 < x*config.TileSize {
			continue
		}
		if y0 < y*config.TileSize {
			return true
		}
		if y1 >= (y+config.PipeGapY)*config.TileSize {
			return true
		}
	}
	return false
}

func (g *Game) drawTiles(screen *ebiten.Image) {
	const (
		nx           = config.ScreenWidth / config.TileSize
		ny           = config.ScreenHeight / config.TileSize
		pipeTileSrcX = 128
		pipeTileSrcY = 192
	)

	op := &ebiten.DrawImageOptions{}
	for i := -2; i < nx+1; i++ {
		// ground
		op.GeoM.Reset()
		op.GeoM.Translate(float64(i*config.TileSize-floorMod(g.cameraX, config.TileSize)),
			float64((ny-1)*config.TileSize-floorMod(g.cameraY, config.TileSize)))
		screen.DrawImage(tilesImage.SubImage(image.Rect(0, 0, config.TileSize, config.TileSize)).(*ebiten.Image), op)

		// pipe
		if tileY, ok := g.pipeAt(floorDiv(g.cameraX, config.TileSize) + i); ok {
			for j := 0; j < tileY; j++ {
				op.GeoM.Reset()
				op.GeoM.Scale(1, -1)
				op.GeoM.Translate(float64(i*config.TileSize-floorMod(g.cameraX, config.TileSize)),
					float64(j*config.TileSize-floorMod(g.cameraY, config.TileSize)))
				op.GeoM.Translate(0, config.TileSize)
				var r image.Rectangle
				if j == tileY-1 {
					r = image.Rect(pipeTileSrcX, pipeTileSrcY, pipeTileSrcX+config.TileSize*2, pipeTileSrcY+config.TileSize)
				} else {
					r = image.Rect(pipeTileSrcX, pipeTileSrcY+config.TileSize, pipeTileSrcX+config.TileSize*2, pipeTileSrcY+config.TileSize*2)
				}
				screen.DrawImage(tilesImage.SubImage(r).(*ebiten.Image), op)
			}
			for j := tileY + config.PipeGapY; j < config.ScreenHeight/config.TileSize-1; j++ {
				op.GeoM.Reset()
				op.GeoM.Translate(float64(i*config.TileSize-floorMod(g.cameraX, config.TileSize)),
					float64(j*config.TileSize-floorMod(g.cameraY, config.TileSize)))
				var r image.Rectangle
				if j == tileY+config.PipeGapY {
					r = image.Rect(pipeTileSrcX, pipeTileSrcY, pipeTileSrcX+config.PipeWidth, pipeTileSrcY+config.TileSize)
				} else {
					r = image.Rect(pipeTileSrcX, pipeTileSrcY+config.TileSize, pipeTileSrcX+config.PipeWidth, pipeTileSrcY+config.TileSize+config.TileSize)
				}
				screen.DrawImage(tilesImage.SubImage(r).(*ebiten.Image), op)
			}
		}
	}
}

func (g *Game) drawGopher(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	w, h := gopherImage.Bounds().Dx(), gopherImage.Bounds().Dy()
	op.GeoM.Translate(-float64(w)/2.0, -float64(h)/2.0)
	op.GeoM.Rotate(float64(g.vy16) / 96.0 * math.Pi / 6)
	op.GeoM.Translate(float64(w)/2.0, float64(h)/2.0)
	op.GeoM.Translate(float64(g.x16/16.0)-float64(g.cameraX), float64(g.y16/16.0)-float64(g.cameraY))
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(gopherImage, op)
}

func main() {
	flag.Parse()
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle("Flappy Gopher With Standings")
	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}
