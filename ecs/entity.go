package ecs

import (
	"sync"
	"sync/atomic"
)

type Entity uint64

var lastEntity = uint64(0)

type ComponentName string

type Component interface {
	Name() ComponentName
	Entity() Entity
}

var entityLock = sync.RWMutex{}
var entitiesByComponent = make(map[ComponentName][]Entity)
var componentsByEntity = make(map[Entity][]Component)

func EntitiesWith(cns ...ComponentName) []Entity {
	// Grab and auto-release the read lock
	entityLock.RLock()
	defer entityLock.RUnlock()

	reply := make([]Entity, 0)
	for _, cn := range cns {
		if ent, ok := entitiesByComponent[cn]; ok {
			reply = append(reply, ent...)
		}
	}

	return reply
}

type EntityBuilder struct {
	entity     Entity
	components []Component
}

func (e *EntityBuilder) With(c ...Component) *EntityBuilder {
	e.components = append(e.components, c...)
	return e
}

func (e *EntityBuilder) Build() Entity {
	// Grab and auto-release the write lock
	entityLock.Lock()
	defer entityLock.Unlock()

	// Add this entity to the list
	componentsByEntity[e.entity] = e.components
	// For every component on this entity
	for _, c := range e.components {
		// See if we already have a list of entities
		if cl, ok := entitiesByComponent[c.Name()]; ok {
			// Append to it if we do
			cl = append(cl, e.entity)
		} else {
			// Create it if we don't
			entitiesByComponent[c.Name()] = []Entity{e.entity}
		}
	}

	return e.entity
}

func NewEntity() *EntityBuilder {
	eid := atomic.AddUint64(&lastEntity, 1)

	return &EntityBuilder{
		entity:     Entity(eid),
		components: make([]Component, 0),
	}
}

func Destroy(e Entity) {
	// Grab the write lock
	entityLock.Lock()
	// And release it when the function completes
	defer entityLock.Unlock()

	// We only have to do work if there exists a list of component names for this entity
	if cbe, ok := componentsByEntity[e]; ok {
		// Create a wait group
		wg := sync.WaitGroup{}
		// And increment it by the number of components on this entity
		wg.Add(len(cbe))
		// For each component
		for _, c := range cbe {
			// Start a goroutine
			go func() {
				// Get the slice of entities that have this component
				eList := entitiesByComponent[c.Name()]
				// Find the destroyed entity
				for i, e2 := range eList {
					if e2 == e {
						// Once we have the destroyed entity, copy the last entity in the slice into its place
						eList[i] = eList[len(eList)-1]
						// Then reduce the size of the slice by one
						entitiesByComponent[c.Name()] = eList[:len(eList)-1]
						// And exit the loop
						break
					}
				}
				// Mark our work as done
				wg.Done()
			}()
		}

		// Wait for all the goroutines to finish
		wg.Wait()
	}
}

func AddComponents(e Entity, c ...Component) {
	entityLock.RLock()
	cl, ok := componentsByEntity[e]
	entityLock.RUnlock()

	if !ok {
		cl = make([]Component, len(c))
		copy(cl, c)
		entityLock.Lock()
		componentsByEntity[e] = cl
	} else {
		cl = append(cl, c...)
		entityLock.Lock()
		componentsByEntity[e] = cl
	}

	entityLock.Unlock()
}

func RemoveComponents(e Entity, cn ...ComponentName) {
	entityLock.Lock()
	defer entityLock.Unlock()

	if cList, ok := componentsByEntity[e]; ok {
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			lastIndex := len(cList) - 1
			for i, c := range cList {
				for _, cn2 := range cn {
					if cn2 == c.Name() {
						cList[i] = cList[lastIndex]
						cList = cList[:lastIndex]
						lastIndex--
					}
				}
			}
			wg.Done()
		}()

		for _, cn2 := range cn {
			myName := cn2
			wg.Add(1)
			go func() {
				eList := entitiesByComponent[myName]
				for i, e2 := range eList {
					if e2 == e {
						eList[i] = eList[len(eList)-1]
						eList = eList[:len(eList)-1]
						entitiesByComponent[myName] = eList
						break
					}
				}

				wg.Done()
			}()
		}

		wg.Wait()
	}
}
