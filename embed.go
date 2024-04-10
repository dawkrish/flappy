package main

import (
	_ "embed"
)

var (
	//go:embed jab.wav
	Jab_wav []byte

	//go:embed jump.ogg
	Jump_ogg []byte

	//go:embed gopher.png
	Gopher_png []byte

	//go:embed background.png
	Background_png []byte
)
