package systems

import (
    "github.com/faiface/pixel"
    "github.com/fogleman/gg"
    "github.com/kintar/gopheroids/ecs"
    "github.com/kintar/gopheroids/util"
    "golang.org/x/image/colornames"
    "math"
)

const EighthCircle = math.Pi / 4.0

func makeAsteroidPicture(radius int) *pixel.PictureData {
    ctx := gg.NewContext(radius, radius)
    ctx.SetColor(colornames.White)
    ctx.MoveTo(1,0)
    for rad := 0.0; rad < util.Tau; rad += EighthCircle {
        x := math.Sin(rad) * float64(radius)
        y := math.Cos(rad) * float64(radius)
        ctx.LineTo(x, y)
    }
    ctx.ClosePath()
    return pixel.PictureDataFromImage(ctx.Image())
}

type Asteroid struct {
    Size float64
    Spin float64
    Velocity pixel.Vec
    SpriteIndex int
}

type AsteroidSystem struct {
    roids map[ecs.Entity]Asteroid
    spriteSheet *pixel.PictureData
    sprites []*pixel.Sprite
}

const AsteroidSystemName = ecs.SystemName("asteroids")

func (a AsteroidSystem) Name() ecs.SystemName {
    return AsteroidSystemName
}

func (a AsteroidSystem) DependsOn() []ecs.SystemName {
    return nil
}

func (a AsteroidSystem) Update(deltaTime float64) {
    panic("implement me")
}
