package systems

import (
    "github.com/faiface/pixel"
    "github.com/fogleman/gg"
    "github.com/kintar/gopheroids/ecs"
    "github.com/kintar/gopheroids/util"
    "golang.org/x/image/colornames"
    "math"
    "math/rand"
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
    Position pixel.Vec
    Rotation float64
    Velocity pixel.Vec
    SpriteIndex int
}

type AsteroidSystem struct {
    drawTarget pixel.Target
    roids map[ecs.Entity]Asteroid
    spriteSheet *pixel.PictureData
    sprites []*pixel.Sprite
}

func NewAsteroidSystem(drawTarget pixel.Target) AsteroidSystem {
    pic := makeAsteroidPicture(20)
    return AsteroidSystem{
        drawTarget:  drawTarget,
        roids:       make(map[ecs.Entity]Asteroid),
        spriteSheet: pic,
        sprites:     []*pixel.Sprite{
            pixel.NewSprite(pic, pic.Bounds()),
        },
    }
}

const AsteroidSystemName = ecs.SystemName("asteroids")
var vecCenter = pixel.V(0,0)

func (a *AsteroidSystem) Name() ecs.SystemName {
    return AsteroidSystemName
}

func (a *AsteroidSystem) DependsOn() []ecs.SystemName {
    return nil
}

func (a *AsteroidSystem) Update(deltaTime float64) {
    batch := pixel.NewBatch(&pixel.TrianglesData{}, a.spriteSheet)
    for _, r := range a.roids {
        r.Position = r.Position.Add(r.Velocity.Scaled(deltaTime))

        if r.Position.X < 0.0 || r.Position.X > 1024 {
            r.Position.X = 1024 - math.Abs(r.Position.X)
        }

        if r.Position.Y < 0 || r.Position.Y > 768 {
            r.Position.Y = 768 - math.Abs(r.Position.Y)
        }

        r.Rotation = r.Rotation + r.Spin * deltaTime
        a.sprites[r.SpriteIndex].Draw(batch, pixel.IM.Moved(r.Position).Rotated(vecCenter, r.Rotation))
    }
    batch.Draw(a.drawTarget)
}

var lastEntity = uint64(0)

func (a *AsteroidSystem) NewRoid() {
    pos := pixel.V(rand.Float64() * 1024, rand.Float64() * 768)
    dir := rand.Float64() * util.Tau
    speed := rand.Float64() * 50
    vel := pixel.V(speed * math.Cos(dir), speed * math.Sin(dir))
    lastEntity++
    ent := ecs.Entity(lastEntity)
    a.roids[ent] = Asteroid{
        Size:        0,
        Spin:        rand.Float64() * math.Pi,
        Position:    pos,
        Rotation:    rand.Float64() * util.Tau,
        Velocity:    vel,
        SpriteIndex: 0,
    }
}
