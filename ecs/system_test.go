package ecs

import (
    "testing"
)

type TestSystem struct {
    updateCount int
}

func (t *TestSystem) Name() SystemName {
    return "Test_System"
}

var updateCount = 0
func (t *TestSystem) Update(deltaTime float64) {
    t.updateCount++
}

func (t *TestSystem) DependsOn() []SystemName {
    return nil
}

func TestRegister(t *testing.T) {
    sys := TestSystem{}
    Register(&sys)

    if _, ok := systems[sys.Name()]; !ok {
        t.Fatalf("system was not registered")
    }
}

func TestUnregister(t *testing.T) {
    sys := TestSystem{}
    Register(&sys)
    count := len(systems)
    Unregister(sys.Name())
    if len(systems) >= count {
        t.Fatalf("expected systems count to be less than %d, but was %d", count, len(systems))
    }
}

func TestUpdate(t *testing.T) {
    sys := &TestSystem{}
    Register(sys)
    Update(0.1)
    if sys.updateCount == 0 {
        t.Fatalf("system did not update")
    }
}
