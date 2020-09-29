package ecs

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type SystemName string

// A System modifies the entities within the game world
type System interface {
	// Name specifies the name of the system, for debugging and for registration
	Name() SystemName
	// Update causes the system to increment its state by the given amount of time (fractional seconds)
	Update(deltaTime float64)
	// DependsOn returns a slice of System names that must run Update before this system Updates
	DependsOn() []SystemName
}

var systems = make(map[SystemName]System)
var systemsMutex = sync.RWMutex{}

func namesIntersect(s1, s2 []SystemName) bool {
	if len(s2) < len(s1) {
		s2, s1 = s1, s2
	}

	for _, n1 := range s1 {
		for _, n2 := range s2 {
			if n1 == n2 {
				return true
			}
		}
	}

	return false
}

// Register adds a system into the processing loop.
func Register(system System) {
	go func() {
		systemsMutex.Lock()
		defer systemsMutex.Unlock()

		// Validate dependencies to prevent cyclic references
		// To do this, we just check if the dependency is already in place.  If not, return an error.
		// This way, cyclic dependencies can never be added due to a chicken-and-egg problem
		// (But it does mean you have to control how you add your systems)
		deps := system.DependsOn()
		for _, dep := range deps {
			if _, ok := systems[dep]; !ok {
				panic(fmt.Sprintf("can't register system %s: missing dependency %s", system.Name(), dep))
			}
		}

		if _, ok := systems[system.Name()]; !ok {
			systems[system.Name()] = system
		} else {
			log.Println("could not add system: system '", system.Name(), "' already registered")
		}
	}()
}

// Unregister removes a System.  Note that it DOES NOT CHECK DEPENDENCIES!  So don't remove a system that another one
// depends on.
func Unregister(system SystemName) {
	go func() {
		systemsMutex.Lock()
		defer systemsMutex.Unlock()

		delete(systems, system)
	}()
}

var updateComplete = make(chan bool)
var lastUpdateDuration float64

// LastUpdateDuration reports the total time elapsed (in seconds) of the last call to Update
func LastUpdateDuration() float64 {
	return lastUpdateDuration
}

// Update processes all registered systems in parallel, based on dependencies
//
// New thought process to allow dependencies between systems:
// 1. Create a map[SystemName]bool to track which ones have and have not completed
// 2. Create wait group
// 3. For each system not yet run:
//   a. Check if the system's dependencies have run
//   b. If so, add one to the wait group
//   c. Start a goroutine that:
//     1. Updates the system
//     2. Notifies the wait group
// 4. Wait for wait group
// 5. If some systems have not yet completed, loop to 3
//
// TODO : Should we put in a guard to prevent a system from calling Update during an update loop?
func Update(deltaTime float64) {
	// Track our update speed
	start := time.Now()

	// Grab the read lock and make a working copy of the current systems
	systemsMutex.RLock()

	completed := make(map[SystemName]bool, len(systems))
	currentSystems := make([]System, len(systems))[:0]
	for name, sys := range systems {
		completed[name] = false
		currentSystems = append(currentSystems, sys)
	}

	// Done with the read lock
	systemsMutex.RUnlock()

	// Create a slice for the pending systems to run on each iteration
	toRun := make([]System, 0)

	// Create a wait group so we can wait on everything in toRun to complete
	waitGroup := sync.WaitGroup{}

	// Loop until we finish updating all of our systems
	for {
		// Truncate our toRun slice
		toRun = toRun[:0]

		// Loop over our current systems
	sysLoop:
		for _, sys := range currentSystems {
			for _, dep := range sys.DependsOn() {
				if !completed[dep] {
					// If any dependency is not complete, skip this system
					continue sysLoop
				}
			}
			// Dependencies are all complete, so we can run it
			toRun = append(toRun, sys)
		}

		// If there's nothing to run, we're finished updating
		if len(toRun) == 0 {
			break
		}

		// Now that we no longer have to read from the 'completed' map, run all of our collected systems in parallel
		for _, sys := range toRun {
			// Capture a local variable for sys
			mySys := sys
			// Increment the wait group
			waitGroup.Add(1)
			// Update the system and signal the group on a new goroutine
			go func() {
				mySys.Update(deltaTime)
				waitGroup.Done()
			}()
		}

		// Wait for all pending systems to update
		waitGroup.Wait()
	}

	lastUpdateDuration = float64(time.Now().Sub(start).Milliseconds()) / 1000.0
}
