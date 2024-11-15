package common

type Object struct {
	// The gopher's position
	X16  int
	Y16  int
	Vy16 int

	// Pipes
	PipeTileYs []int
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
