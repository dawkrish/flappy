package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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

	vy     float64
	dy     float64
	height int
	width  int
	jump   float64
}

type Pipe struct {
	img  *ebiten.Image
	opts ebiten.DrawImageOptions
	rect image.Rectangle

	height  int
	crossed bool
}

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeOver

	screenWidth      = 800
	screenHeight     = 600
	pipeWidth        = 70
	pipeDx           = 3.2
	numberOfPipePair = 5
	pipeGap          = 270
)

var (
	goph             Player
	bg1              BackgroundImage
	bg2              BackgroundImage
	pipes            [][2]Pipe
	arcadeFaceSource *text.GoTextFaceSource
)

// Constants only
func init() {
	img, _, err := image.Decode(bytes.NewReader(Gopher_png))
	if err != nil {
		log.Fatal(err)
	}
	goph.img = ebiten.NewImageFromImage(img)

	goph.vy = 0.15
	goph.jump = 55
	goph.width = goph.img.Bounds().Dx()
	goph.height = goph.img.Bounds().Dy()
	goph.opts.GeoM.Translate(float64(screenWidth-goph.width)/2, float64(screenHeight-goph.height)/2)
}

func init() {
	img, _, err := image.Decode(bytes.NewReader(Background_png))
	if err != nil {
		log.Fatal(err)
	}
	backgroundImg := ebiten.NewImageFromImage(img)
	imgX := float64(backgroundImg.Bounds().Dx())
	imgY := float64(backgroundImg.Bounds().Dy())
	ratioX := float64(screenWidth) / imgX
	ratioY := float64(screenHeight) / imgY

	bg1.img, bg2.img = backgroundImg, backgroundImg
	bg1.dx, bg2.dx = 4, 4
	bg1.opts.GeoM.Scale(ratioX, ratioY)
	bg2.opts.GeoM.Scale(ratioX, ratioY)
}

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf))
	if err != nil {
		log.Fatal(err)
	}
	arcadeFaceSource = s
}

// this struct implements ebiten.Game interface
type Game struct {
	score int
	count int
	mode  Mode

	audioContext *audio.Context
	jumpPlayer   *audio.Player
	hitPlayer    *audio.Player

	BottomBoundary float64
	TopBoundary    float64
}

func NewGame() *Game {
	return &Game{
		count:          0,
		score:          0,
		mode:           ModeTitle,
		BottomBoundary: 500,
		TopBoundary:    0,
	}
}

func (g *Game) init() {
	g.count = 0
	g.score = 0
	SetXcord(&bg1.opts, 0)
	SetXcord(&bg2.opts, screenWidth)
	SetYcord(&goph.opts, float64(screenHeight-goph.height)/2)

	goph.dy = 1.5
	goph.rect.Min = image.Point{(screenWidth - goph.width) / 2, (screenHeight - goph.height) / 2}
	goph.rect.Max = image.Point{(screenWidth + goph.width) / 2, (screenHeight + goph.height) / 2}

	pipes = populatePipes()

	if g.audioContext == nil {
		g.audioContext = audio.NewContext(48000)
	}

	jumpD, err := vorbis.DecodeWithoutResampling(bytes.NewReader(Jump_ogg))
	if err != nil {
		log.Fatal(err)
	}

	g.jumpPlayer, err = g.audioContext.NewPlayer(jumpD)
	if err != nil {
		log.Fatal(err)
	}

	jabD, err := wav.DecodeWithoutResampling(bytes.NewReader(Jab_wav))
	if err != nil {
		log.Fatal(err)
	}
	g.hitPlayer, err = g.audioContext.NewPlayer(jabD)
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Update() error {
	switch g.mode {
	case ModeTitle:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.init()
			g.mode = ModeGame
		}

	case ModeGame:
		g.count++
		// Infinite background scroll
		bg1.opts.GeoM.Translate(-bg1.dx, 0)
		bg2.opts.GeoM.Translate(-bg2.dx, 0)

		if GetXcord(bg1.opts)+float64(screenWidth) <= 0 {
			SetXcord(&bg1.opts, float64(screenWidth))
		}
		if GetXcord(bg2.opts)+float64(screenWidth) <= 0 {
			SetXcord(&bg2.opts, float64(screenWidth))
		}

		// Gopher
		goph.dy += goph.vy
		goph.opts.GeoM.Translate(0, goph.dy)

		// As we do not want continuous jump
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
			if err := g.jumpPlayer.Rewind(); err != nil {
				return err
			}
			g.jumpPlayer.Play()
			goph.dy = 0
			goph.opts.GeoM.Translate(0, -goph.jump)
		}
		goph.rect.Min.Y = int(GetYcord(goph.opts))
		goph.rect.Max.Y = int(GetYcord(goph.opts)) + goph.height

		// Detect Collision
		// Ground
		if GetYcord(goph.opts)+float64(goph.height) >= screenHeight {
			g.hitPlayer.Play()
			g.mode = ModeOver
		}
		// Top
		if GetYcord(goph.opts) < -100 {
			g.hitPlayer.Play()
			g.mode = ModeOver
		}

		// Pipe
		for i := 0; i < len(pipes); i++ {
			needToSwitchHeights := false
			for j := 0; j < 2; j++ {
				pipes[i][j].opts.GeoM.Translate(-pipeDx, 0)
				pipes[i][j].rect = image.Rect(int(GetXcord(pipes[i][j].opts)), int(GetYcord(pipes[i][j].opts)), int(GetXcord(pipes[i][j].opts)+pipeWidth), int(GetYcord(pipes[i][j].opts))+pipes[i][j].height)

				if GetXcord(pipes[i][j].opts)+float64(pipeWidth) <= 0 {
					needToSwitchHeights = true
					previousPipePairIndex := (i + 4) % 5
					newXcord := GetXcord(pipes[previousPipePairIndex][0].opts) + pipeGap
					SetXcord(&pipes[i][j].opts, newXcord)
					pipes[i][j].rect.Min.X = int(GetXcord(pipes[i][j].opts))
					pipes[i][j].rect.Max.X = int(GetXcord(pipes[i][j].opts)) + pipeWidth
					pipes[i][j].crossed = false
				}
				if pipes[i][j].rect.Overlaps(goph.rect) {
					g.hitPlayer.Play()
					g.mode = ModeOver
				}
				if j == 0 && GetXcord(pipes[i][j].opts)+float64(pipeWidth) <= GetXcord(goph.opts)-float64(goph.width) && !pipes[i][j].crossed {
					g.score++
					pipes[i][j].crossed = true
					// log.Println(g.score)
				}
			}

			if needToSwitchHeights {
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

				pipes[i][0].img = img1
				pipes[i][1].img = img2
				SetYcord(&pipes[i][1].opts, float64(screenHeight-height2))
				pipes[i][0].rect.Min.Y = 0
				pipes[i][0].rect.Max.Y = height1
				pipes[i][1].rect.Min.Y = int(GetYcord(pipes[i][1].opts))
				pipes[i][1].rect.Max.Y = int(GetYcord(pipes[i][1].opts)) + height2
			}
		}
	case ModeOver:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			time.Sleep(1 * time.Second)
			g.init()
			g.mode = ModeGame
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.mode {
	case ModeTitle:
		screen.DrawImage(bg1.img, &bg1.opts)
		op := &text.DrawOptions{}
		op.GeoM.Translate(screenWidth/2, screenHeight/2)
		op.ColorScale.ScaleWithColor(color.White)
		op.PrimaryAlign = text.AlignCenter
		text.Draw(screen, "Press Space/Enter to start the game", &text.GoTextFace{
			Source: arcadeFaceSource,
			Size:   20,
		}, op)
		op.GeoM.Translate(0, 50)
		text.Draw(screen, "Press Space/LeftClick to jump", &text.GoTextFace{
			Source: arcadeFaceSource,
			Size:   20,
		}, op)

	case ModeGame:
		screen.DrawImage(bg1.img, &bg1.opts)
		screen.DrawImage(bg2.img, &bg2.opts)
		for _, pipeR := range pipes {
			for _, pipe := range pipeR {
				screen.DrawImage(pipe.img, &pipe.opts)
			}
		}
		screen.DrawImage(goph.img, &goph.opts)
		op := &text.DrawOptions{}
		op.GeoM.Translate(screenWidth/2, 100)
		op.ColorScale.ScaleWithColor(color.White)
		op.PrimaryAlign = text.AlignCenter
		textMsg := fmt.Sprintf("%v", g.score)
		text.Draw(screen, textMsg, &text.GoTextFace{
			Source: arcadeFaceSource,
			Size:   50,
		}, op)
	case ModeOver:
		screen.DrawImage(bg1.img, &bg1.opts)
		screen.DrawImage(bg2.img, &bg2.opts)
		for _, pipeR := range pipes {
			for _, pipe := range pipeR {
				screen.DrawImage(pipe.img, &pipe.opts)
			}
		}
		screen.DrawImage(goph.img, &goph.opts)
		op := &text.DrawOptions{}
		op.GeoM.Translate(screenWidth/2, 100)
		op.ColorScale.ScaleWithColor(color.White)
		op.PrimaryAlign = text.AlignCenter
		textMsg := fmt.Sprintf("Final Score : %v", g.score)
		text.Draw(screen, textMsg, &text.GoTextFace{
			Source: arcadeFaceSource,
			Size:   30,
		}, op)
		op.GeoM.Translate(0, 200)
		text.Draw(screen, "Press Enter to restart the game", &text.GoTextFace{
			Source: arcadeFaceSource,
			Size:   20,
		}, op)

	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	g := NewGame()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("flappy gopher")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
