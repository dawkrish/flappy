package main

import (
	"image"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

func GetXcord(opts ebiten.DrawImageOptions) float64 {
	return opts.GeoM.Element(0, 2)
}
func GetYcord(opts ebiten.DrawImageOptions) float64 {
	return opts.GeoM.Element(1, 2)
}

func SetXcord(opts *ebiten.DrawImageOptions, xCord float64) {
	opts.GeoM.SetElement(0, 2, xCord)
}
func SetYcord(opts *ebiten.DrawImageOptions, yCord float64) {
	opts.GeoM.SetElement(1, 2, yCord)
}

func (goph *Player) UpdateRect() {
	goph.rect.Min.Y = int(GetYcord(goph.opts))
	goph.rect.Max.Y = int(GetYcord(goph.opts)) + goph.height
}

func (pipe *Pipe) UpdateRect() {
	minX := int(GetXcord(pipe.opts))
	minY := int(GetYcord(pipe.opts))
	maxX := int(GetXcord(pipe.opts) + pipeWidth)
	maxY := int(GetYcord(pipe.opts) + float64(pipe.height))
	pipe.rect = image.Rect(minX, minY, maxX, maxY)

}

func ChangeHeights(pipe1, pipe2 Pipe) {
	var height1 int
	var height2 int
	var minimumHeight = 80
	var totalHeight = 425

	if rand.Float64() > 0.5 {
		height1 = minimumHeight + rand.Intn(75)
		height2 = totalHeight - height1
	} else {
		height2 = minimumHeight + rand.Intn(75)
		height1 = totalHeight - height2
	}
	img1 := ebiten.NewImage(pipeWidth, height1)
	img1.Fill(color.RGBA{30, 200, 15, 0xff})
	img2 := ebiten.NewImage(pipeWidth, height2)
	img2.Fill(color.RGBA{30, 200, 15, 0xff})

	pipe1.img = img1
	pipe2.img = img2

	pipe1.height = height1
	pipe2.height = height2
}

func PopulatePipes() [][2]Pipe {
	var pipes [][2]Pipe
	for i := 0; i < 5; i++ {
		var height1 int
		var height2 int
		var minimumHeight = 80
		var totalHeight = 425

		if rand.Float64() > 0.5 {
			height1 = minimumHeight + rand.Intn(75)
			height2 = totalHeight - height1
		} else {
			height2 = minimumHeight + rand.Intn(75)
			height1 = totalHeight - height2
		}
		img1 := ebiten.NewImage(pipeWidth, height1)
		img1.Fill(color.RGBA{30, 200, 15, 0xff})
		img2 := ebiten.NewImage(pipeWidth, height2)
		img2.Fill(color.RGBA{30, 200, 15, 0xff})

		opts1 := ebiten.DrawImageOptions{}
		opts2 := ebiten.DrawImageOptions{}

		offsetX := screenWidth
		xCord := float64(offsetX + (pipeWidth+pipeGap)*i)

		opts1.GeoM.Translate(xCord, 0)
		opts2.GeoM.Translate(xCord, float64(screenHeight-height2))
		p1 := Pipe{
			img:     img1,
			opts:    opts1,
			rect:    image.Rect(int(GetXcord(opts1)), int(GetYcord(opts1)), int(GetXcord(opts1)+pipeWidth), int(GetYcord(opts1))+height1),
			height:  height1,
			crossed: false,
		}
		p2 := Pipe{
			img:     img2,
			opts:    opts2,
			rect:    image.Rect(int(GetXcord(opts2)), int(GetYcord(opts2)), int(GetXcord(opts2)+pipeWidth), int(GetYcord(opts2))+height2),
			height:  height2,
			crossed: false,
		}
		pRow := [2]Pipe{p1, p2}
		pipes = append(pipes, pRow)
	}
	return pipes
}
