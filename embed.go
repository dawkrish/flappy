package main

import (
	_ "embed"
)

var (
	//go:embed assets/jab.wav
	Jab_wav []byte

	//go:embed assets/jump.ogg
	Jump_ogg []byte

	//go:embed assets/gopher.png
	Gopher_png []byte

	//go:embed assets/background.png
	Background_png []byte
)
