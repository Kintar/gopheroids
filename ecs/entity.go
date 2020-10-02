package ecs

import (
    "sync"
    "sync/atomic"
)

type EntityId uint64

type ComponentId uint64

type Entity struct {
    EntityId
    manager    *EntityManager
    components sync.Map
    destroyed  bool
}

func (e *Entity) Add(cid ComponentId, c interface{}) {
    e.components.Store(cid, c)
    e.manager.registerComponent(cid, e)
}

func (e *Entity) Destroy() {
    e.manager.DestroyEntity(e.EntityId)
}

type EntityManager struct {
    lastEntityId        uint64
    lastComponentId     uint64
    entities            sync.Map
    entitiesByComponent map[ComponentId][]*Entity
    ebcMutex            sync.RWMutex
}

func (m *EntityManager) registerComponent(cid ComponentId, e *Entity) {
    m.ebcMutex.Lock()
    defer m.ebcMutex.Unlock()

    var eList []*Entity

    if el, ok := m.entitiesByComponent[cid]; ok {
        eList = append(el, e)
    } else {
        eList = []*Entity{e}
    }

    m.entitiesByComponent[cid] = eList
}

func (m *EntityManager) removeComponent(cid ComponentId, e *Entity) {
    m.ebcMutex.Lock()
    defer m.ebcMutex.Unlock()

    if el, ok := m.entitiesByComponent[cid]; ok {
        for i, e2 := range el {
            if e == e2 {
                el[i] = el[len(el)-1]
                el = el[:len(el)-1]
                m.entitiesByComponent[cid] = el
                return
            }
        }
    }
}

func (m *EntityManager) NewEntity() *Entity {
    e := &Entity{
        EntityId:   EntityId(atomic.AddUint64(&m.lastEntityId, 1)),
        manager:    m,
        components: sync.Map{},
    }

    m.entities.Store(e.EntityId, e)
    return e
}

func (m *EntityManager) CreateComponent() ComponentId {
    return ComponentId(atomic.AddUint64(&m.lastComponentId, 1))
}

func (m *EntityManager) DestroyEntity(eid EntityId) {
    e, ok := m.entities.LoadAndDelete(eid)
    if !ok {
        return
    }

    ent := e.(*Entity)
    ent.components = sync.Map{}
    ent.manager = nil
    ent.destroyed = true
    ent.EntityId = 0
}

func (m *EntityManager) Query(cids... ComponentId) []*Entity {
    m.ebcMutex.RLock()
    defer m.ebcMutex.Unlock()

    entities := make([]*Entity, 0)
    for _, cid := range cids {
        if es, ok := m.entitiesByComponent[cid]; ok {
            for _, e := range es {
                if e.destroyed {
                    continue
                }
                entities = append(entities, e)
            }
        }
    }
    return entities
}
