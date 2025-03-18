package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// ボタン構造体
type Button struct {
	X, Y          int
	Width, Height int
	Text          string
	FontSize      float64
	Enabled       bool
}

func newButton(x, y, width, height int, text string, fontSize float64) Button {
	return Button{
		X:        x,
		Y:        y,
		Width:    width,
		Height:   height,
		Text:     text,
		FontSize: fontSize,
		Enabled:  true,
	}
}

// ボタンが押されたかチェック
func (b *Button) IsClicked() bool {
	if !b.Enabled {
		return false
	}

	// マウスクリックのチェック
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if x >= b.X && x <= b.X+b.Width && y >= b.Y && y <= b.Y+b.Height {
			return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
		}
	}

	// タッチのチェック
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

// ボタンの描画
func (b *Button) Draw(screen *ebiten.Image) {
	if !b.Enabled {
		return
	}

	// ボタンの背景
	bgRect := ebiten.NewImage(b.Width, b.Height)
	bgRect.Fill(color.RGBA{0x40, 0x40, 0x60, 0xff})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.X), float64(b.Y))
	screen.DrawImage(bgRect, op)

	// ボタンの枠線（上）
	topBorder := ebiten.NewImage(b.Width, 2)
	topBorder.Fill(color.White)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.X), float64(b.Y))
	screen.DrawImage(topBorder, op)

	// ボタンの枠線（左）
	leftBorder := ebiten.NewImage(2, b.Height)
	leftBorder.Fill(color.White)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.X), float64(b.Y))
	screen.DrawImage(leftBorder, op)

	// ボタンの枠線（下）
	bottomBorder := ebiten.NewImage(b.Width, 2)
	bottomBorder.Fill(color.White)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.X), float64(b.Y+b.Height-2))
	screen.DrawImage(bottomBorder, op)

	// ボタンの枠線（右）
	rightBorder := ebiten.NewImage(2, b.Height)
	rightBorder.Fill(color.White)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.X+b.Width-2), float64(b.Y))
	screen.DrawImage(rightBorder, op)

	// ボタンのテキスト
	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(float64(b.X+b.Width/2), float64(b.Y+b.Height/2))
	textOp.ColorScale.ScaleWithColor(color.White)
	textOp.PrimaryAlign = text.AlignCenter
	// 垂直方向の中央揃えのために、Y座標を調整
	textOp.GeoM.Translate(0, -b.FontSize/3) // フォントサイズに応じて調整
	text.Draw(screen, b.Text, &text.GoTextFace{
		Source: arcadeFaceSource,
		Size:   b.FontSize,
	}, textOp)
}
