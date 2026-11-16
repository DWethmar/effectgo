# effectgo

Particle effects playground for [Ebitengine](https://ebitengine.org): confetti, stars, shapes and a "ball pop" burst.

```
go run .
```

- Left click: play the current effect
- Right click / Space: next effect
- 1-9: pick an effect

Every effect is plain data (`Effect` with `Options` layers) in `main.go`.
`effects.go` is the self-contained engine: copy it into another game, keep a `System`,
call `Update`/`Draw` each frame and `Play(&effect, x, y)` to trigger one.
All particles are drawn in a single batched draw call.
