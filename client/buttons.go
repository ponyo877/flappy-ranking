package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Button struct {
	X, Y          int
	Width, Height int
	Text          string
	FontSize      float64
	bgColor       color.Color
	Enabled       bool
}

func newButton(x, y, width, height int, text string, fontSize float64, bgColor color.Color) Button {
	return Button{
		X:        x,
		Y:        y,
		Width:    width,
		Height:   height,
		Text:     text,
		FontSize: fontSize,
		bgColor:  bgColor,
		Enabled:  true,
	}
}

func (b *Button) IsClicked() bool {
	if !b.Enabled {
		return false
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if x >= b.X && x <= b.X+b.Width && y >= b.Y && y <= b.Y+b.Height {
			return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
		}
	}
	touchIDs := []ebiten.TouchID{}
	touchIDs = ebiten.AppendTouchIDs(touchIDs[:0])
	for _, touchID := range touchIDs {
		x, y := ebiten.TouchPosition(touchID)
		if x >= b.X && x <= b.X+b.Width && y >= b.Y && y <= b.Y+b.Height {
			return inpututil.IsTouchJustReleased(touchID)
		}
	}
	return false
}

func (b *Button) Draw(screen *ebiten.Image) {
	if !b.Enabled {
		return
	}

	bgRect := ebiten.NewImage(b.Width, b.Height)
	bgRect.Fill(b.bgColor)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.X), float64(b.Y))
	screen.DrawImage(bgRect, op)

	topBorder := ebiten.NewImage(b.Width, 2)
	topBorder.Fill(color.White)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.X), float64(b.Y))
	screen.DrawImage(topBorder, op)

	leftBorder := ebiten.NewImage(2, b.Height)
	leftBorder.Fill(color.White)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.X), float64(b.Y))
	screen.DrawImage(leftBorder, op)

	bottomBorder := ebiten.NewImage(b.Width, 2)
	bottomBorder.Fill(color.White)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.X), float64(b.Y+b.Height-2))
	screen.DrawImage(bottomBorder, op)

	rightBorder := ebiten.NewImage(2, b.Height)
	rightBorder.Fill(color.White)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.X+b.Width-2), float64(b.Y))
	screen.DrawImage(rightBorder, op)

	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(float64(b.X+b.Width/2), float64(b.Y+b.Height/2))
	textOp.ColorScale.ScaleWithColor(color.White)
	textOp.PrimaryAlign = text.AlignCenter

	textOp.GeoM.Translate(0, -b.FontSize/3)
	text.Draw(screen, b.Text, &text.GoTextFace{
		Source: arcadeFaceSource,
		Size:   b.FontSize,
	}, textOp)
}
