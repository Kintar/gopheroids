package ecs

import "sync/atomic"

type Updatable interface {
    Update(deltaTime float64)
}

type Component interface {
    Updatable
    Id() ComponentId
}

type ComponentId uint64

type ComponentRegistry struct {
    lastComponentId uint64
}

func (c *ComponentRegistry) CreateComponent() ComponentId {
    return ComponentId(atomic.AddUint64(&c.lastComponentId, 1))
}
