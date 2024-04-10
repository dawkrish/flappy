package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type BackgroundImage struct {
	img  *ebiten.Image
	opts ebiten.DrawImageOptions
	dx   float64
}

type Player struct {
	img  *ebiten.Image
	opts ebiten.DrawImageOptions
	rect image.Rectangle

	vy float64
	dy float64

	height int
	width  int
	jump   float64
}

type Pipe struct {
	img  *ebiten.Image
	opts ebiten.DrawImageOptions
	rect image.Rectangle

	height             int
	crossedMiddlePoint bool
}

type Mode int
