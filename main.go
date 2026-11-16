// effectgo: a playground for particle effects in Ebitengine.
// Left click: play the current effect. Right click / Space: next effect. 1-9: pick effect.
// Every effect is plain data below: tweak the Options to change how it looks.
package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const screenW, screenH = 960, 640

var (
	party = []color.RGBA{{255, 89, 94, 255}, {255, 202, 58, 255}, {138, 201, 38, 255},
		{25, 130, 196, 255}, {106, 76, 147, 255}, {255, 146, 76, 255}, {240, 240, 255, 255}}
	gold  = []color.RGBA{{255, 220, 90, 255}, {255, 240, 170, 255}}
	white = []color.RGBA{{240, 240, 255, 255}}
)

var (
	confetti = Options{Count: 120, Kinds: []Kind{Rect}, Colors: party,
		Angle: -math.Pi / 2, Spread: 1.8, SpeedMin: 4, SpeedMax: 11, SizeMin: 6, SizeMax: 11,
		LifeMin: 90, LifeMax: 150, Gravity: 0.18, Friction: 0.03, Spin: 0.3, Flutter: 0.08,
		SpawnRadius: 10, ScaleIn: 6, FadeOut: 0.4}
	stars = Options{Count: 24, Kinds: []Kind{Star}, Colors: gold,
		Spread: 2 * math.Pi, SpeedMin: 1, SpeedMax: 5, SizeMin: 6, SizeMax: 14,
		LifeMin: 50, LifeMax: 90, Gravity: 0.03, Friction: 0.05, Spin: 0.1, Twinkle: 0.25,
		SpawnRadius: 12, ScaleIn: 8, FadeOut: 0.4}
	sparkles = Options{Count: 40, Kinds: []Kind{Circle}, Colors: white,
		Spread: 2 * math.Pi, SpeedMin: 0.5, SpeedMax: 3, SizeMin: 1, SizeMax: 2.5,
		LifeMin: 30, LifeMax: 60, Friction: 0.04, SpawnRadius: 8, FadeOut: 0.5}
	shapes = Options{Count: 50, Kinds: []Kind{Tri, Circle, Rect, Star}, Colors: party,
		Spread: 2 * math.Pi, SpeedMin: 2, SpeedMax: 8, SizeMin: 5, SizeMax: 10,
		LifeMin: 60, LifeMax: 110, Gravity: 0.12, Friction: 0.03, Spin: 0.2,
		SpawnRadius: 10, ScaleIn: 6, FadeOut: 0.4}
	shockwave = Options{Count: 1, Kinds: []Kind{Ring}, Colors: white,
		SizeMin: 10, SizeMax: 10, LifeMin: 22, LifeMax: 22, Grow: 4, FadeOut: 1}
)

// withCount returns a copy of o with a different particle count.
func withCount(o Options, n int) Options { o.Count = n; return o }

var effects = []Effect{
	{Name: "Ball pop (marble)",
		Ball:   &BallOptions{Radius: 18, Color: color.RGBA{60, 140, 230, 255}, Shine: color.RGBA{220, 240, 255, 255}, Ticks: 20, Swell: 0.25},
		Layers: []Options{shockwave, stars, sparkles, withCount(shapes, 60)}},
	{Name: "Confetti", Layers: []Options{confetti}},
	{Name: "Stars", Layers: []Options{stars, sparkles}},
	{Name: "Shapes", Layers: []Options{shapes}},
}

type Game struct {
	fx     System
	effect int
}

func (g *Game) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.fx.Play(&effects[g.effect], float64(x), float64(y))
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.effect = (g.effect + 1) % len(effects)
	}
	for i := range min(len(effects), 9) {
		if inpututil.IsKeyJustPressed(ebiten.Key1 + ebiten.Key(i)) {
			g.effect = i
		}
	}
	g.fx.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{18, 18, 28, 255})
	g.fx.Draw(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Effect %d/%d: %s\nLeft click: play   Right click/Space: next   1-%d: pick\nParticles: %d  FPS: %.0f",
		g.effect+1, len(effects), effects[g.effect].Name, len(effects), g.fx.Len(), ebiten.ActualFPS()))
}

func (g *Game) Layout(_, _ int) (int, int) { return screenW, screenH }

func main() {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("effectgo")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
