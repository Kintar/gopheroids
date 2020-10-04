// Package ecs provides a simple but robust entity component system.
//
// Entities provide no game functionality, but are simply ways to group pieces of functionality together on a game
// object.  Components provide the logic to update game state on each frame.  Systems can be used to extend game logic
// in cases where logic on a single component doesn't make sense or becomes a potential race between components.
//
// The main game loop must call EntityManager.Update once every frame.  The EntityManager will process Update on all
// Component instances in parallel, then on all System instances in dependency order, in parallel when possible.
package ecs
