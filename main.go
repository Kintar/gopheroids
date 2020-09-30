package main

import (
    "github.com/faiface/pixel"
    "github.com/faiface/pixel/pixelgl"
    "github.com/kintar/gopheroids/ecs"
    "github.com/kintar/gopheroids/systems"
    "golang.org/x/image/colornames"
    "time"
)

func main() {
    pixelgl.Run(run)
}

func run() {
    window, err := pixelgl.NewWindow(pixelgl.WindowConfig{
        Title:     "Gopheroids!",
        Bounds:    pixel.Rect{Min: pixel.V(0,0), Max: pixel.V(1024, 768)},
        Resizable: false,
        VSync:     true,
    })

    if err != nil {
        panic(err)
    }

    window.Show()

    roids := systems.NewAsteroidSystem(window)

    for r := 0; r < 20; r++ {
        roids.NewRoid()
    }

    ecs.Register(&roids)

    last := time.Now()
    for ; !window.Closed(); {
        window.Clear(colornames.Black)
        now := time.Now()
        delta := float64(now.Sub(last).Milliseconds()) / 1000.0
        ecs.Update(delta)
        if window.JustPressed(pixelgl.KeyEscape) {
            window.SetClosed(true)
        }
        window.Update()
    }
}
