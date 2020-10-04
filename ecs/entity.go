package ecs

import (
    "sync"
    "sync/atomic"
)

type EntityId uint64

type Entity struct {
    EntityId
    components []Component
    manager    EntityManager
}

func (e *Entity) Update(deltaTime float64) {
    for _, c := range e.components {
        c.Update(deltaTime)
    }
}

func (e *Entity) Add(id ComponentId, component interface{}) {
    e.manager.Add(e.EntityId, id, component)
}

func (e *Entity) Remove(id ComponentId) {
    e.manager.Remove(e.EntityId, id)
}

func (e *Entity) Destroy() {
    e.manager.Destroy(e.EntityId)
}

type EntityListener interface {
    Added(EntityId, ComponentId, interface{})
    Removed(EntityId, ComponentId)
    Destroyed(EntityId)
}

type EntityManager interface {
    Updatable
    RegisterInterest(ComponentId, EntityListener)
    RemoveInterest(ComponentId, EntityListener)
    Create() *Entity
    Destroy(EntityId)
    Add(EntityId, ComponentId, interface{})
    Remove(EntityId, ComponentId)
}

type entityManager struct {
    lastEntityId uint64
    entities map[EntityId]*Entity
    systems []System
    entityListeners map[ComponentId][]EntityListener
    updateMutex sync.Mutex
}

func NewEntityManager() EntityManager {
    return &entityManager{
        entities:        make(map[EntityId]*Entity),
        systems:         make([]System, 0),
        entityListeners: make(map[ComponentId][]EntityListener),
    }
}

func (e *entityManager) RegisterInterest(id ComponentId, listener EntityListener) {
    e.updateMutex.Lock()
    defer e.updateMutex.Unlock()

    slice, ok := e.entityListeners[id]
    if !ok {
        slice = make([]EntityListener, 1)
        slice[0] = listener
    } else {
        slice = append(slice, listener)
    }

    e.entityListeners[id] = slice
}

func (e *entityManager) RemoveInterest(id ComponentId, listener EntityListener) {
    e.updateMutex.Lock()
    defer e.updateMutex.Unlock()

    slice, ok := e.entityListeners[id]
    if ok {
        end := len(slice) - 1
        for i, el := range slice {
            if el == listener {
                slice[i] = slice[end]
                slice = slice[:end]
                e.entityListeners[id] = slice
                return
            }
        }
    }
}

func (e *entityManager) Create() *Entity {
    eid := atomic.AddUint64(&e.lastEntityId, 1)
    entity := &Entity{
        EntityId: EntityId(eid),
        manager: e,
        components: make([]Component, 0),
    }
    e.updateMutex.Lock()
    e.entities[entity.EntityId] = entity
    e.updateMutex.Unlock()
    return entity
}

func (e *entityManager) Destroy(id EntityId) {
    e.updateMutex.Lock()
    delete(e.entities, id)
    e.updateMutex.Unlock()
}

func (e *entityManager) Add(id EntityId, id2 ComponentId, i interface{}) {
    panic("implement me")
}

func (e *entityManager) Remove(id EntityId, id2 ComponentId) {
    panic("implement me")
}

func (e *entityManager) Update(deltaTime float64) {
    panic("implement me")
}
