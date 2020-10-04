// Package ecs provides a simple but robust entity component system.
//
// Component implementors are responsible for per-frame updates of their own state.
// System implementors are responsible for updating state where coordination between
// multiple Entity or Component instances is necessary
package ecs
