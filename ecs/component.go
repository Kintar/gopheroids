package ecs

type Updatable interface {
    Update(deltaTime float64)
}

type Component interface {
    Updatable
    Id() ComponentId
}

type ComponentId uint64

const NoComponent = ComponentId(0)
