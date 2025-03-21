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
	"net/url"

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
	"github.com/ponyo877/flappy-ranking/common"
)

var (
	gopherImage      *ebiten.Image
	tilesImage       *ebiten.Image
	arcadeFaceSource *text.GoTextFaceSource
	host             *url.URL
	serverEndpoint   = "http://localhost:8787"
	buttonColor1     = color.RGBA{0x60, 0x80, 0xa0, 0xff}
	buttonColor2     = color.RGBA{0x60, 0x60, 0x80, 0xff}
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
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf))
	if err != nil {
		log.Fatal(err)
	}
	arcadeFaceSource = s
	host, err = url.Parse(serverEndpoint)
	if err != nil {
		log.Fatal(err)
	}
}

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
	ModeRanking
)

type Game struct {
	mode Mode
	obj  *common.Object

	// Camera
	cameraX int
	cameraY int

	gameoverCount int

	touchIDs   []ebiten.TouchID
	gamepadIDs []ebiten.GamepadID

	audioContext *audio.Context
	jumpPlayer   *audio.Player
	hitPlayer    *audio.Player

	jumpHistory []int

	token        string
	pipeKey      string
	playerName   string
	errorMessage string

	// スコア送信成功フラグ
	scoreSubmitted bool

	rankings        []*common.Score
	rankingPeriod   string // "DAILY", "WEEKLY", "MONTHLY"
	fetchingRanking bool

	rankingButton     Button
	dailyButton       Button
	weeklyButton      Button
	monthlyButton     Button
	backButton        Button
	submitScoreButton Button
}

func NewGame() ebiten.Game {
	g := &Game{}
	g.init()
	return g
}

func (g *Game) init() {
	g.cameraX = common.InitialCameraX
	g.cameraY = common.InitialCameraY

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
	g.scoreSubmitted = false

	g.rankingButton = newButton(
		common.ScreenWidth/2-80,
		common.ScreenHeight-100,
		160,
		40,
		"RANKING",
		common.MiddleFontSize,
		buttonColor1,
	)

	buttonWidth := 100
	buttonHeight := 40
	buttonY := common.ScreenHeight - 80
	buttonSpacing := 20

	g.dailyButton = newButton(
		50,
		buttonY,
		buttonWidth,
		buttonHeight,
		"DAILY",
		common.SmallFontSize,
		buttonColor2,
	)

	g.weeklyButton = newButton(
		50+buttonWidth+buttonSpacing,
		buttonY,
		buttonWidth,
		buttonHeight,
		"WEEKLY",
		common.SmallFontSize,
		buttonColor2,
	)

	g.monthlyButton = newButton(
		50+2*(buttonWidth+buttonSpacing),
		buttonY,
		buttonWidth,
		buttonHeight,
		"MONTHLY",
		common.SmallFontSize,
		buttonColor2,
	)

	g.backButton = newButton(
		50+3*(buttonWidth+buttonSpacing),
		buttonY,
		buttonWidth,
		buttonHeight,
		"BACK",
		common.SmallFontSize,
		buttonColor2,
	)

	g.submitScoreButton = newButton(
		common.ScreenWidth/2-80,
		common.ScreenHeight-250,
		160,
		40,
		"SUBMIT",
		common.MiddleFontSize,
		buttonColor1,
	)
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
	return common.ScreenWidth, common.ScreenHeight
}

func (g *Game) Update() error {
	switch g.mode {
	case ModeTitle:
		if g.rankingButton.IsClicked() || inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.rankingPeriod = "DAILY"
			go g.fetchRanking()
			g.mode = ModeRanking
			return nil
		}

		if g.isKeyJustPressed() {
			g.fetchToken()
			g.obj = common.NewObject(
				common.InitialX16,
				common.InitialY16,
				0,
				g.pipeKey,
			)
			g.mode = ModeGame
		}
	case ModeGame:
		g.obj.X16 += common.DeltaX16
		g.cameraX += common.DeltaCameraX
		if g.isKeyJustPressed() {
			g.jumpHistory = append(g.jumpHistory, g.obj.X16)
			g.obj.Vy16 = -common.VyLimit
			if err := g.jumpPlayer.Rewind(); err != nil {
				return err
			}
			g.jumpPlayer.Play()
		}
		g.obj.Y16 += g.obj.Vy16

		// Gravity
		g.obj.Vy16 += common.DeltaVy16
		if g.obj.Vy16 > common.VyLimit {
			g.obj.Vy16 = common.VyLimit
		}

		if g.obj.Hit() {
			// log.Printf("debug jumpHistory: %v", g.jumpHistory)
			if err := g.hitPlayer.Rewind(); err != nil {
				return err
			}
			g.hitPlayer.Play()
			g.mode = ModeGameOver
			g.gameoverCount = 0
			g.finishSession()
		}
	case ModeGameOver:
		g.gameoverCount++

		// Input
		runes := ebiten.AppendInputChars(nil)
		for _, r := range runes {
			if r != ' ' && len(g.playerName) < 10 {
				g.playerName += string(r)
			}
		}

		// Delete
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.playerName) > 0 {
			g.playerName = g.playerName[:len(g.playerName)-1]
		}

		playerName := g.playerName
		if len(g.playerName) == 0 {
			playerName = "NO NAME"
		}
		if g.submitScoreButton.IsClicked() || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.submitScore(playerName)
			return nil
		}
		if g.gameoverCount > 30 && g.isKeyJustPressed() {
			g.init()
			g.mode = ModeTitle
		}
	case ModeRanking:
		if g.dailyButton.IsClicked() || inpututil.IsKeyJustPressed(ebiten.KeyD) {
			g.rankingPeriod = "DAILY"
			go g.fetchRanking()
		}
		if g.weeklyButton.IsClicked() || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			g.rankingPeriod = "WEEKLY"
			go g.fetchRanking()
		}
		if g.monthlyButton.IsClicked() || inpututil.IsKeyJustPressed(ebiten.KeyM) {
			g.rankingPeriod = "MONTHLY"
			go g.fetchRanking()
		}
		if g.backButton.IsClicked() || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.mode = ModeTitle
		}
	}
	return nil
}

func (g *Game) drawRanking(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x40, 0x40, 0x60, 0xff})

	title := fmt.Sprintf("%s RANKING", g.rankingPeriod)
	op := &text.DrawOptions{}
	op.GeoM.Translate(common.ScreenWidth/2, 50)
	op.ColorScale.ScaleWithColor(color.White)
	op.PrimaryAlign = text.AlignCenter
	text.Draw(screen, title, &text.GoTextFace{
		Source: arcadeFaceSource,
		Size:   common.TitleFontSize,
	}, op)
	// ランキング表示
	if g.fetchingRanking {
		op.GeoM.Translate(0, 100)
		text.Draw(screen, "Loading...", &text.GoTextFace{
			Source: arcadeFaceSource,
			Size:   common.FontSize,
		}, op)
	} else if len(g.rankings) == 0 {
		op.GeoM.Translate(0, 100)
		text.Draw(screen, "No data", &text.GoTextFace{
			Source: arcadeFaceSource,
			Size:   common.FontSize,
		}, op)
	} else {
		for i, score := range g.rankings {
			// Top 10
			if i > 9 {
				break
			}

			y := 100 + i*30
			rankText := fmt.Sprintf("%2d. %-10s %4d", score.Rank, score.DisplayName, score.Score)

			op := &text.DrawOptions{}
			op.GeoM.Translate(common.ScreenWidth/2, float64(y))
			op.ColorScale.ScaleWithColor(color.White)
			op.PrimaryAlign = text.AlignCenter
			text.Draw(screen, rankText, &text.GoTextFace{
				Source: arcadeFaceSource,
				Size:   common.FontSize,
			}, op)
		}
	}

	g.dailyButton.Draw(screen)
	g.weeklyButton.Draw(screen)
	g.monthlyButton.Draw(screen)
	g.backButton.Draw(screen)

	guide := "D: Daily  W: Weekly  M: Monthly  ESC: Back"
	op = &text.DrawOptions{}
	op.GeoM.Translate(common.ScreenWidth/2, common.ScreenHeight-30)
	op.ColorScale.ScaleWithColor(color.White)
	op.PrimaryAlign = text.AlignCenter
	text.Draw(screen, guide, &text.GoTextFace{
		Source: arcadeFaceSource,
		Size:   common.SmallFontSize,
	}, op)
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.mode == ModeRanking {
		g.drawRanking(screen)
		return
	}

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
		texts = "\n\n\n\nPRESS SPACE KEY\n\nOR A/B BUTTON\n\nOR TOUCH SCREEN\n\nPRESS R FOR RANKING"
		g.rankingButton.Draw(screen)
	case ModeGameOver:
		if g.scoreSubmitted {
			texts = "\nSCORE SUBMITTED!\n\n\n\n\n\n\n\nPRESS KEY TO CONTINUE"
		} else {
			cursor := " "
			if g.gameoverCount%30 < 15 {
				cursor = "_"
			}
			texts = "\nENTER OR SUBMIT YOUR NAME:\n\n" + g.playerName + cursor + "\n\n\n\n\n\nPRESS KEY TO CONTINUE"
			g.submitScoreButton.Draw(screen)
		}
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(common.ScreenWidth/2, 3*common.TitleFontSize)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = common.TitleFontSize
	op.PrimaryAlign = text.AlignCenter
	text.Draw(screen, titleTexts, &text.GoTextFace{
		Source: arcadeFaceSource,
		Size:   common.TitleFontSize,
	}, op)

	op = &text.DrawOptions{}
	op.GeoM.Translate(common.ScreenWidth/2, 3*common.TitleFontSize)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = common.FontSize
	op.PrimaryAlign = text.AlignCenter
	text.Draw(screen, texts, &text.GoTextFace{
		Source: arcadeFaceSource,
		Size:   common.FontSize,
	}, op)

	if g.mode == ModeTitle {
		const msg = "Go Gopher by Renee French is\nlicenced under CC BY 3.0."

		op := &text.DrawOptions{}
		op.GeoM.Translate(common.ScreenWidth/2, common.ScreenHeight-common.SmallFontSize/2)
		op.ColorScale.ScaleWithColor(color.White)
		op.LineSpacing = common.SmallFontSize
		op.PrimaryAlign = text.AlignCenter
		op.SecondaryAlign = text.AlignEnd
		text.Draw(screen, msg, &text.GoTextFace{
			Source: arcadeFaceSource,
			Size:   common.SmallFontSize,
		}, op)
	}

	op = &text.DrawOptions{}
	op.GeoM.Translate(common.ScreenWidth, 0)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = common.FontSize
	op.PrimaryAlign = text.AlignEnd
	if g.obj != nil {
		text.Draw(screen, fmt.Sprintf("%04d", g.obj.Score()), &text.GoTextFace{
			Source: arcadeFaceSource,
			Size:   common.FontSize,
		}, op)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func (g *Game) drawTiles(screen *ebiten.Image) {
	const (
		nx           = common.ScreenWidth / common.TileSize
		ny           = common.ScreenHeight / common.TileSize
		pipeTileSrcX = 128
		pipeTileSrcY = 192
	)

	op := &ebiten.DrawImageOptions{}
	for i := -2; i < nx+1; i++ {
		// ground
		op.GeoM.Reset()
		op.GeoM.Translate(float64(i*common.TileSize-common.FloorMod(g.cameraX, common.TileSize)),
			float64((ny-1)*common.TileSize-common.FloorMod(g.cameraY, common.TileSize)))
		screen.DrawImage(tilesImage.SubImage(image.Rect(0, 0, common.TileSize, common.TileSize)).(*ebiten.Image), op)

		// pipe
		if tileY, ok := g.obj.PipeAt(common.FloorDiv(g.cameraX, common.TileSize) + i); ok {
			for j := 0; j < tileY; j++ {
				op.GeoM.Reset()
				op.GeoM.Scale(1, -1)
				op.GeoM.Translate(float64(i*common.TileSize-common.FloorMod(g.cameraX, common.TileSize)),
					float64(j*common.TileSize-common.FloorMod(g.cameraY, common.TileSize)))
				op.GeoM.Translate(0, common.TileSize)
				var r image.Rectangle
				if j == tileY-1 {
					r = image.Rect(pipeTileSrcX, pipeTileSrcY, pipeTileSrcX+common.TileSize*2, pipeTileSrcY+common.TileSize)
				} else {
					r = image.Rect(pipeTileSrcX, pipeTileSrcY+common.TileSize, pipeTileSrcX+common.TileSize*2, pipeTileSrcY+common.TileSize*2)
				}
				screen.DrawImage(tilesImage.SubImage(r).(*ebiten.Image), op)
			}
			for j := tileY + common.PipeGapY; j < common.ScreenHeight/common.TileSize-1; j++ {
				op.GeoM.Reset()
				op.GeoM.Translate(float64(i*common.TileSize-common.FloorMod(g.cameraX, common.TileSize)),
					float64(j*common.TileSize-common.FloorMod(g.cameraY, common.TileSize)))
				var r image.Rectangle
				if j == tileY+common.PipeGapY {
					r = image.Rect(pipeTileSrcX, pipeTileSrcY, pipeTileSrcX+common.PipeWidth, pipeTileSrcY+common.TileSize)
				} else {
					r = image.Rect(pipeTileSrcX, pipeTileSrcY+common.TileSize, pipeTileSrcX+common.PipeWidth, pipeTileSrcY+common.TileSize+common.TileSize)
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
	op.GeoM.Rotate(float64(g.obj.Vy16) / 96.0 * math.Pi / 6)
	op.GeoM.Translate(float64(w)/2.0, float64(h)/2.0)
	op.GeoM.Translate(float64(g.obj.X16/16.0)-float64(g.cameraX), float64(g.obj.Y16/16.0)-float64(g.cameraY))
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(gopherImage, op)
}

func main() {
	flag.Parse()
	ebiten.SetTPS(60)
	ebiten.SetWindowSize(common.ScreenWidth, common.ScreenHeight)
	ebiten.SetWindowTitle("Flappy Gopher With Ranking")
	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}
