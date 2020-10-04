package ecs

import (
    "log"
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
    CreateComponent() ComponentId
    RegisterInterest(ComponentId, EntityListener)
    RemoveInterest(ComponentId, EntityListener)
    Create() *Entity
    Destroy(EntityId)
    Add(EntityId, ComponentId, interface{})
    Remove(EntityId, ComponentId)
}

type entityManager struct {
    lastEntityId    uint64
    lastComponentId uint64
    entities        map[EntityId]*Entity
    systems         []System
    entityListeners map[ComponentId][]EntityListener
    updateMutex     sync.Mutex
}

func NewEntityManager() EntityManager {
    return &entityManager{
        entities:        make(map[EntityId]*Entity),
        systems:         make([]System, 0),
        entityListeners: make(map[ComponentId][]EntityListener),
    }
}

func (e *entityManager) CreateComponent() ComponentId {
    return ComponentId(atomic.AddUint64(&e.lastComponentId, 1))
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
        EntityId:   EntityId(eid),
        manager:    e,
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

func (e *entityManager) notify(id ComponentId, callback func (EntityListener)) {
    e.updateMutex.Lock()
    defer e.updateMutex.Unlock()

    if ls, ok := e.entityListeners[id]; ok {
        for _, l := range ls {
            callback(l)
        }
    }
}

func (e *entityManager) Add(id EntityId, id2 ComponentId, i interface{}) {
    if component, ok := i.(Component); !ok {
        log.Println("attempted to add non-component to entity")
    } else {
        if ent, ok := e.entities[id]; ok {
            ent.components = append(ent.components, component)

            go e.notify(id2, func (el EntityListener) {
                el.Added(id, id2, component)
            })
        }
    }
}

func (e *entityManager) Remove(id EntityId, id2 ComponentId) {
    if ent, ok := e.entities[id]; ok {
        end := len(ent.components) - 1
        for i, c := range ent.components {
            if c.Id() == id2 {
                e.updateMutex.Lock()
                ent.components[i] = ent.components[end]
                ent.components = ent.components[:end]
                e.updateMutex.Unlock()
                go e.notify(id2, func (listener EntityListener) {
                    listener.Removed(id, id2)
                })
            }
        }
    }
}

func (e *entityManager) Update(deltaTime float64) {
    wg := sync.WaitGroup{}
    e.updateMutex.Lock()
    wg.Add(len(e.entities))
    for _, ent := range e.entities {
        self := ent
        go func() {
            self.Update(deltaTime)
            wg.Done()
        }()
    }
    e.updateMutex.Unlock()
    wg.Wait()
}
