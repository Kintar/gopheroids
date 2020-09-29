package ecs

import (
	"sync"
	"sync/atomic"
)

type Entity uint64

type Component uint64

type ComponentManager interface {
	SetComponentType(Component)
	AddToEntity(Entity)
	RemoveFromEntity(Entity)
}

type EntityManager interface {
	RegisterComponentManager(ComponentManager)
	CreateEntity() Entity
	DestroyEntity(Entity)
	ComponentsForEntity(Entity) []Component
	EntitiesWithComponents(types ...Component) []Entity
}

type entityManager struct {
	lastEntity          uint64
	lastComponent       Component
	componentManagers   map[Component]ComponentManager
	componentsByEntity  map[Entity][]Component
	entitiesByComponent map[Component][]Entity
	componentMutex      sync.RWMutex
}

func NewEntityManager() EntityManager {
	return &entityManager{
		lastEntity:          0,
		lastComponent:       0,
		componentManagers:   make(map[Component]ComponentManager),
		componentsByEntity:  make(map[Entity][]Component),
		entitiesByComponent: make(map[Component][]Entity),
		componentMutex:      sync.RWMutex{},
	}
}

func (e *entityManager) RegisterComponentManager(manager ComponentManager) {
	e.componentMutex.Lock()

	e.lastComponent++
	cType := e.lastComponent
	e.componentManagers[cType] = manager

	e.componentMutex.Unlock()

	manager.SetComponentType(cType)
}

func (e *entityManager) CreateEntity() Entity {
	eid := atomic.AddUint64(&(e.lastEntity), 1)
	return Entity(eid)
}

func (e *entityManager) DestroyEntity(entity Entity) {
	e.componentMutex.RLock()
	components, ok := e.componentsByEntity[entity]
	e.componentMutex.RUnlock()

	if !ok {
		return
	}

	e.componentMutex.Lock()
	delete(e.componentsByEntity, entity)
	e.componentMutex.Unlock()

	for _, c := range components {
		e.componentManagers[c].RemoveFromEntity(entity)
	}
}

func (e *entityManager) ComponentsForEntity(entity Entity) []Component {
	e.componentMutex.RLock()
	defer e.componentMutex.RUnlock()
	if c, ok := e.componentsByEntity[entity]; ok {
		return c
	}
	return make([]Component, 0)
}

func (e *entityManager) EntitiesWithComponents(types ...Component) []Entity {
	e.componentMutex.RLock()
	defer e.componentMutex.RUnlock()
	entities := make([]Entity, 0)
	for _, t := range types {
		if ent, ok := e.entitiesByComponent[t]; ok {
			entities = append(entities, ent...)
		}
	}
	return entities
}
