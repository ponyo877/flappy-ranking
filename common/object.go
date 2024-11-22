package common

import (
	"math/rand/v2"
	"time"
)

type Object struct {
	// The gopher's position
	X16  int
	Y16  int
	Vy16 int

	// Pipes
	PipeTileYs []int
}

func NewObject(initX16, initY16, initVy16 int, pipeKey string) *Object {
	pipeTileYs := make([]int, 256)
	seed := [32]byte([]byte(pipeKey))
	r := rand.New(rand.NewChaCha8(seed))
	for i := range pipeTileYs {
		pipeTileYs[i] = r.IntN(6) + 2
	}
	return &Object{
		X16:        initX16,
		Y16:        initY16,
		Vy16:       initVy16,
		PipeTileYs: pipeTileYs,
	}
}

func (o *Object) PipeAt(tileX int) (tileY int, ok bool) {
	if (tileX - PipeStartOffsetX) <= 0 {
		return 0, false
	}
	if FloorMod(tileX-PipeStartOffsetX, PipeIntervalX) != 0 {
		return 0, false
	}
	idx := FloorDiv(tileX-PipeStartOffsetX, PipeIntervalX)
	return o.PipeTileYs[idx%len(o.PipeTileYs)], true
}

func (o *Object) Score() int {
	x := FloorDiv(o.X16, Unit) / TileSize
	if (x - PipeStartOffsetX) <= 0 {
		return 0
	}
	return FloorDiv(x-PipeStartOffsetX, PipeIntervalX)
}

func (o *Object) Hit() bool {
	const (
		gopherWidth  = 30
		gopherHeight = 60
	)
	// w, h := gopherImage.Bounds().Dx(), gopherImage.Bounds().Dy()
	w, h := 60, 75

	x0 := FloorDiv(o.X16, Unit) + (w-gopherWidth)/2
	y0 := FloorDiv(o.Y16, Unit) + (h-gopherHeight)/2
	x1 := x0 + gopherWidth
	y1 := y0 + gopherHeight
	if y0 < -TileSize*4 {
		return true
	}
	if y1 >= ScreenHeight-TileSize {
		return true
	}
	xMin := FloorDiv(x0-PipeWidth, TileSize)
	xMax := FloorDiv(x0+gopherWidth, TileSize)
	for x := xMin; x <= xMax; x++ {
		y, ok := o.PipeAt(x)
		if !ok {
			continue
		}
		if x0 >= x*TileSize+PipeWidth {
			continue
		}
		if x1 < x*TileSize {
			continue
		}
		if y0 < y*TileSize {
			return true
		}
		if y1 >= (y+PipeGapY)*TileSize {
			return true
		}
	}
	return false
}

func (o *Object) IsValidTimeDiff(startTime, endTime time.Time) bool {
	diffSecond := int(endTime.Sub(startTime).Seconds())
	gameSec60FPS := o.X16 / DeltaX16 / 60
	intervalSec60FPS := PipeIntervalX / DeltaX16 / 60
	// Between 30FPS and 60FPS is valid
	return gameSec60FPS <= diffSecond && diffSecond <= (gameSec60FPS+intervalSec60FPS)*2
}
