package main

import (
	"image"
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Kind int

const (
	Rect   Kind = iota // confetti strip with a fake 3D flip
	Star               // 5-point star
	Circle             // dot
	Tri                // triangle
	Ring               // outline circle, e.g. a shockwave (use Grow)
)

// Options controls one layer of particles. The zero value of every field means "off".
// Durations are in ticks (60 ticks = 1 second), angles in radians.
type Options struct {
	Count              int
	Kinds              []Kind       // picked at random per particle; empty = Circle
	Colors             []color.RGBA // picked at random per particle; empty = white
	Angle, Spread      float64      // emit direction (0 = right, -π/2 = up) and cone width (2π = all around)
	SpeedMin, SpeedMax float64      // pixels per tick
	SizeMin, SizeMax   float64      // pixels
	LifeMin, LifeMax   int
	Gravity            float64 // added to vertical speed each tick
	Friction           float64 // 0..1, fraction of speed lost each tick
	Spin               float64 // max rotation speed per tick
	Flutter            float64 // sideways wobble, makes confetti float
	Twinkle            float64 // 0..1 size pulse amount
	Grow               float64 // size added per tick
	SpawnRadius        float64 // particles start up to this far out along their direction
	ScaleIn            int     // ticks to grow from nothing to full size
	FadeOut            float64 // 0..1, last fraction of life spent fading out
}

// BallOptions makes an effect start with a ball that swells, shrinks away and then bursts.
type BallOptions struct {
	Radius       float64
	Color, Shine color.RGBA
	Ticks        int     // total time until the burst
	Swell        float64 // how much bigger it gets before shrinking (0.25 = 25%)
}

// Effect is a named set of particle layers, optionally preceded by a ball.
type Effect struct {
	Name   string
	Ball   *BallOptions
	Layers []Options
}

type particle struct {
	x, y, vx, vy, rot, vrot, size float64
	life, maxLife                 int
	col                           color.RGBA
	k                             Kind
	o                             *Options
}

type activeBall struct {
	x, y float64
	t    int
	e    *Effect
}

// System runs and draws effects. Copy effects.go into another game to reuse it.
type System struct {
	ps    []particle
	balls []activeBall
	vs    []ebiten.Vertex // reused each frame
	is    []uint16
}

func randf(a, b float64) float64 { return a + rand.Float64()*(b-a) }

func (s *System) Len() int { return len(s.ps) + len(s.balls) }

// Play starts an effect at (x, y).
func (s *System) Play(e *Effect, x, y float64) {
	if e.Ball != nil {
		s.balls = append(s.balls, activeBall{x: x, y: y, e: e})
		return
	}
	for i := range e.Layers {
		s.Emit(e.Layers[i], x, y)
	}
}

// Emit spawns one layer of particles at (x, y).
func (s *System) Emit(o Options, x, y float64) {
	op := &o
	for range o.Count {
		a := o.Angle + randf(-o.Spread/2, o.Spread/2)
		sp := randf(o.SpeedMin, o.SpeedMax)
		cos, sin := math.Cos(a), math.Sin(a)
		r := rand.Float64() * o.SpawnRadius
		p := particle{
			x: x + cos*r, y: y + sin*r, vx: cos * sp, vy: sin * sp,
			rot: randf(0, 2*math.Pi), vrot: randf(-o.Spin, o.Spin),
			size: randf(o.SizeMin, o.SizeMax),
			life: max(1, int(randf(float64(o.LifeMin), float64(o.LifeMax)))),
			col:  color.RGBA{255, 255, 255, 255}, k: Circle, o: op,
		}
		p.maxLife = p.life
		if len(o.Colors) > 0 {
			p.col = o.Colors[rand.IntN(len(o.Colors))]
		}
		if len(o.Kinds) > 0 {
			p.k = o.Kinds[rand.IntN(len(o.Kinds))]
		}
		s.ps = append(s.ps, p)
	}
}

// Update advances everything one tick.
func (s *System) Update() {
	balls := s.balls[:0]
	for _, b := range s.balls {
		b.t++
		if b.t >= b.e.Ball.Ticks {
			for i := range b.e.Layers {
				s.Emit(b.e.Layers[i], b.x, b.y)
			}
			continue
		}
		balls = append(balls, b)
	}
	s.balls = balls

	ps := s.ps[:0]
	for _, p := range s.ps {
		o := p.o
		p.vy += o.Gravity
		p.vx *= 1 - o.Friction
		p.vy *= 1 - o.Friction
		p.vx += math.Sin(float64(p.life)*0.15+p.rot) * o.Flutter
		p.size += o.Grow
		p.x += p.vx
		p.y += p.vy
		p.rot += p.vrot
		p.life--
		if p.life > 0 {
			ps = append(ps, p)
		}
	}
	s.ps = ps
}

// ballRadius: quick pop-in to full swell in the first quarter, then shrink to nothing.
func ballRadius(b *BallOptions, t int) float64 {
	f := float64(t) / float64(b.Ticks)
	peak := b.Radius * (1 + b.Swell)
	if f < 0.25 {
		return peak * math.Sin(f/0.25*math.Pi/2)
	}
	g := (f - 0.25) / 0.75
	return peak * (1 - g*g)
}

func (s *System) Draw(dst *ebiten.Image) {
	for _, b := range s.balls {
		r := ballRadius(b.e.Ball, b.t)
		vector.DrawFilledCircle(dst, float32(b.x), float32(b.y), float32(r), b.e.Ball.Color, true)
		vector.DrawFilledCircle(dst, float32(b.x-r*0.35), float32(b.y-r*0.35), float32(r*0.3), b.e.Ball.Shine, true)
	}

	// All filled particles go into one vertex buffer and one draw call; per-particle draws were the bottleneck.
	s.vs, s.is = s.vs[:0], s.is[:0]
	for _, p := range s.ps {
		o := p.o
		age := p.maxLife - p.life
		a := 1.0
		if o.FadeOut > 0 {
			a = math.Min(1, float64(p.life)/(float64(p.maxLife)*o.FadeOut))
		}
		sz := p.size
		if o.ScaleIn > 0 {
			sz *= math.Min(1, float64(age)/float64(o.ScaleIn))
		}
		sz *= 1 - o.Twinkle*(0.5+0.5*math.Sin(float64(p.life)*0.5))
		c := fade(p.col, a)
		switch p.k {
		case Rect:
			s.addFan(p.x, p.y, p.rot, sz/2, sz*0.3*math.Abs(math.Cos(p.rot*1.7)), unitRect, c)
		case Star:
			s.addFan(p.x, p.y, p.rot, sz, sz, unitStar, c)
		case Tri:
			s.addFan(p.x, p.y, p.rot, sz, sz, unitTri, c)
		case Circle:
			s.addFan(p.x, p.y, 0, sz/2, sz/2, unitCircle, c)
		case Ring:
			vector.StrokeCircle(dst, float32(p.x), float32(p.y), float32(sz), float32(4*a), c, true)
		}
		if len(s.vs) > 65000 { // uint16 index limit
			s.flush(dst)
		}
	}
	s.flush(dst)
}

func (s *System) flush(dst *ebiten.Image) {
	if len(s.is) > 0 {
		dst.DrawTriangles(s.vs, s.is, whiteSub, &ebiten.DrawTrianglesOptions{ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha, AntiAlias: true})
	}
	s.vs, s.is = s.vs[:0], s.is[:0]
}

// addFan appends shape (unit points around the origin) scaled by sx, sy, rotated and moved to (cx, cy).
// A fan from the center works because every shape here is star-shaped around its center.
func (s *System) addFan(cx, cy, rot, sx, sy float64, shape [][2]float64, c color.RGBA) {
	sin, cos := math.Sincos(rot)
	r, g, b, a := float32(c.R)/255, float32(c.G)/255, float32(c.B)/255, float32(c.A)/255
	base := uint16(len(s.vs))
	s.vs = append(s.vs, ebiten.Vertex{DstX: float32(cx), DstY: float32(cy), SrcX: 1, SrcY: 1, ColorR: r, ColorG: g, ColorB: b, ColorA: a})
	for _, p := range shape {
		x, y := p[0]*sx, p[1]*sy
		s.vs = append(s.vs, ebiten.Vertex{DstX: float32(cx + x*cos - y*sin), DstY: float32(cy + x*sin + y*cos), SrcX: 1, SrcY: 1, ColorR: r, ColorG: g, ColorB: b, ColorA: a})
	}
	n := uint16(len(shape))
	for i := range n {
		s.is = append(s.is, base, base+1+i, base+1+(i+1)%n)
	}
}

var whiteSub = func() *ebiten.Image {
	img := ebiten.NewImage(3, 3)
	img.Fill(color.White)
	return img.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
}()

// fade returns c premultiplied by alpha a (0..1).
func fade(c color.RGBA, a float64) color.RGBA {
	return color.RGBA{uint8(float64(c.R) * a), uint8(float64(c.G) * a), uint8(float64(c.B) * a), uint8(float64(c.A) * a)}
}

var (
	unitRect   = [][2]float64{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}
	unitTri    = [][2]float64{{0, -1}, {0.87, 0.5}, {-0.87, 0.5}}
	unitStar   = ngon(10, 0.45)
	unitCircle = ngon(12, 1)
)

// ngon returns n points on the unit circle starting at the top; every odd point is at radius inner.
func ngon(n int, inner float64) [][2]float64 {
	pts := make([][2]float64, n)
	for i := range pts {
		r := 1.0
		if i%2 == 1 {
			r = inner
		}
		a := float64(i)*2*math.Pi/float64(n) - math.Pi/2
		pts[i] = [2]float64{math.Cos(a) * r, math.Sin(a) * r}
	}
	return pts
}
