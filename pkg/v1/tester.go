package v1

import (
	"fmt"
	"log"
	"sync"
)

// StageFunc represents the function to be executed in a stage.
type StageFunc func()

// StageDef represents a defined stage.
type StageDef struct {
	Name string
	Func StageFunc
}

// Tester is the main struct for the integration test library.
type Tester struct {
	Stages []StageDef
	mu     sync.Mutex
}

// NewTester creates a new Tester instance.
func NewTester() *Tester {
	return &Tester{
		Stages: make([]StageDef, 0),
	}
}

// Stage registers a new stage.
func (t *Tester) Stage(name string, fn StageFunc) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Stages = append(t.Stages, StageDef{Name: name, Func: fn})
}

// RunStageByName runs a specific stage by name.
func (t *Tester) RunStageByName(name string) (err error) {
	t.mu.Lock()
	var fn StageFunc
	for _, s := range t.Stages {
		if s.Name == name {
			fn = s.Func
			break
		}
	}
	t.mu.Unlock()

	if fn == nil {
		return fmt.Errorf("stage %s not found", name)
	}

	log.Printf("=== Running Stage: %s ===", name)
	// Error handling in stages should be handled by panic/recover or other means if we want to stop execution
	// For this lib, we assume stages might panic on failure.
	defer func() {
		if r := recover(); r != nil {
			log.Printf("=== Stage %s FAILED: %v ===", name, r)
			err = fmt.Errorf("panic: %v", r)
		} else {
			log.Printf("=== Stage %s PASSED ===", name)
		}
	}()
	fn()
	return nil
}
