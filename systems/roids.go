package systems

import (
    "fmt"
    "github.com/faiface/pixel"
    "github.com/fogleman/gg"
    "github.com/kintar/gopheroids/ecs"
    "github.com/kintar/gopheroids/util"
    "golang.org/x/image/colornames"
    "math"
    "math/rand"
    "sync"
)

const CircleSegment = math.Pi / 8.0

func makeAsteroidSpriteSheet() (*pixel.PictureData, []*pixel.Sprite) {
    sprites := make([]*pixel.Sprite, 0)

    ctx := gg.NewContext(240, 400)
    ctx.Identity()
    xOffset := 40.0
    ctx.Translate(xOffset, 40)
    ctx.SetColor(colornames.White)

    spriteBounds := make([]pixel.Rect, 0)

    // Make five asteroid images at each of the three size scales
    for scale := 80.0; scale >= 20.0; scale /= 2 {
        for a := 0; a < 5; a++ {
            yOffset := 80 * float64(a)
            ctx.Translate(0, yOffset)
            drawAsteroid(scale, ctx)
            ox, oy := ctx.TransformPoint(-scale/2, -scale/2)
            ctx.Translate(0, -yOffset)
            spriteBounds = append(spriteBounds, pixel.R(ox, oy, ox+scale, oy+scale))
        }
        ctx.Translate(scale, 0)
    }

    pic := pixel.PictureDataFromImage(ctx.Image())
    for _, bounds := range spriteBounds {
        fmt.Printf("%v\n", bounds)
        sprites = append(sprites, pixel.NewSprite(pic, bounds))
    }

    return pic, sprites
}

func drawAsteroid(radius float64, ctx *gg.Context) {
    ctx.SetColor(colornames.White)

    quarterRadius := radius / 4.0

    angle := 0.0

    offsetX := math.Cos(angle) * quarterRadius
    offsetY := math.Sin(angle) * quarterRadius

    angle += CircleSegment
    for ; angle < util.Tau; angle += CircleSegment {
        offsetX = math.Cos(angle) * (quarterRadius + rand.Float64()*quarterRadius)
        offsetY = math.Sin(angle) * (quarterRadius + rand.Float64()*quarterRadius)
        ctx.LineTo(offsetX, offsetY)
    }
    ctx.ClosePath()
    ctx.Stroke()
}

type Asteroid struct {
    entity      *ecs.Entity
    Size        float64
    Spin        float64
    Position    pixel.Vec
    Rotation    float64
    Velocity    pixel.Vec
    SpriteIndex int
}

func (r *Asteroid) Update(deltaTime float64) {
    r.Position = r.Position.Add(r.Velocity.Scaled(deltaTime))

    if r.Position.X < 0.0 || r.Position.X > 1024 {
        r.Position.X = 1024 - math.Abs(r.Position.X)
    }

    if r.Position.Y < 0 || r.Position.Y > 768 {
        r.Position.Y = 768 - math.Abs(r.Position.Y)
    }

    r.Rotation = r.Rotation + r.Spin*deltaTime
}

func (a Asteroid) Id() ecs.ComponentId {
    return roidComponent
}

var roidComponent ecs.ComponentId

type AsteroidSystem struct {
    updateMutex   sync.RWMutex
    manager       ecs.EntityManager
    drawTarget    pixel.Target
    spriteSheet   *pixel.PictureData
    sprites       []*pixel.Sprite
    roids         map[ecs.EntityId]*Asteroid
}

func NewAsteroidSystem(drawTarget pixel.Target, mgr ecs.EntityManager) *AsteroidSystem {
    pic, sprites := makeAsteroidSpriteSheet()
    roidComponent = mgr.CreateComponent()
    sys := &AsteroidSystem{
        manager:       mgr,
        drawTarget:    drawTarget,
        spriteSheet:   pic,
        sprites:       sprites,
        roids:         make(map[ecs.EntityId]*Asteroid, 0),
    }
    mgr.RegisterInterest(roidComponent, sys)
    return sys
}

const AsteroidSystemName = ecs.SystemName("asteroids")

var vecCenter = pixel.V(0, 0)

func (a *AsteroidSystem) Name() ecs.SystemName {
    return AsteroidSystemName
}

func (a *AsteroidSystem) DependsOn() []ecs.SystemName {
    return nil
}

func (a *AsteroidSystem) Added(id ecs.EntityId, id2 ecs.ComponentId, i interface{}) {
    if id2 != roidComponent {
        return
    }

    if roid, ok := i.(*Asteroid); ok {
        a.updateMutex.Lock()
        defer a.updateMutex.Unlock()

        a.roids[id] = roid
    }
}

func (a *AsteroidSystem) Removed(id ecs.EntityId, id2 ecs.ComponentId) {
    if id2 == roidComponent {
        a.updateMutex.Lock()
        defer a.updateMutex.Unlock()
        delete(a.roids, id)
    }
}

func (a *AsteroidSystem) Destroyed(id ecs.EntityId) {
    a.updateMutex.Lock()
    delete(a.roids, id)
    a.updateMutex.Unlock()
}

func (a *AsteroidSystem) Update(_ float64) {
    batch := pixel.NewBatch(&pixel.TrianglesData{}, a.spriteSheet)
    a.updateMutex.RLock()
    defer a.updateMutex.RUnlock()

    for _, r := range a.roids {
        a.sprites[r.SpriteIndex].Draw(batch, pixel.IM.Rotated(vecCenter, r.Rotation).Moved(r.Position))
    }
    batch.Draw(a.drawTarget)
}

var lastRoid = 0

func (a *AsteroidSystem) NewRoid() {
    pos := pixel.V(rand.Float64()*1024, rand.Float64()*768)
    dir := rand.Float64() * util.Tau
    speed := rand.Float64() * 50
    vel := pixel.V(speed*math.Cos(dir), speed*math.Sin(dir))
    c := &Asteroid{
        Size:        0,
        Spin:        -0.5 + rand.Float64(),
        Position:    pos,
        Rotation:    rand.Float64() * util.Tau,
        Velocity:    vel,
        SpriteIndex: lastRoid % 5,
    }
    lastRoid++
    e := a.manager.Create()
    e.Add(roidComponent, c)
}
