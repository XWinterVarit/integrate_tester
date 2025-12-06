package v1

import (
	"fmt"
	"sync"
)

// StageFunc represents the function to be executed in a stage.
type StageFunc func()

// StageDef represents a defined stage.
type StageDef struct {
	Name string
	Func StageFunc
}

// Action represents a runnable operation within a stage.
type Action struct {
	Summary string
	Func    func()
}

var (
	// stageActions maps StageName -> List of Actions
	stageActions = make(map[string][]Action)
	// currentStage tracks the currently running stage name
	currentStage string
	// isRecording determines if operations should be recorded
	isRecording bool
	// actionMu protects the global state
	actionMu sync.Mutex
	// actionHandlers are notified when actions list updates
	actionHandlers []func()
)

// RecordAction registers an operation for the current stage.
func RecordAction(summary string, fn func()) {
	actionMu.Lock()
	defer actionMu.Unlock()

	if !isRecording || currentStage == "" {
		return
	}

	stageActions[currentStage] = append(stageActions[currentStage], Action{
		Summary: summary,
		Func:    fn,
	})

	notifyActionHandlers()
}

// GetStageActions returns the recorded actions for a stage.
func GetStageActions(stageName string) []Action {
	actionMu.Lock()
	defer actionMu.Unlock()
	// Return copy to be safe
	src := stageActions[stageName]
	dst := make([]Action, len(src))
	copy(dst, src)
	return dst
}

// RegisterActionUpdateHandler adds a listener for action updates.
func RegisterActionUpdateHandler(fn func()) {
	actionMu.Lock()
	defer actionMu.Unlock()
	actionHandlers = append(actionHandlers, fn)
}

func notifyActionHandlers() {
	for _, h := range actionHandlers {
		h()
	}
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

	// Setup context for recording
	actionMu.Lock()
	currentStage = name
	isRecording = true
	stageActions[name] = []Action{} // Clear previous actions
	notifyActionHandlers()
	actionMu.Unlock()

	Log(LogTypeStage, fmt.Sprintf("Running Stage: %s", name), "")

	// Ensure recording stops after stage
	defer func() {
		actionMu.Lock()
		isRecording = false
		currentStage = ""
		actionMu.Unlock()
	}()

	// Error handling in stages should be handled by panic/recover or other means if we want to stop execution
	// For this lib, we assume stages might panic on failure.
	defer func() {
		if r := recover(); r != nil {
			Log(LogTypeStage, fmt.Sprintf("Stage %s FAILED", name), fmt.Sprintf("%v", r))
			err = fmt.Errorf("panic: %v", r)
		} else {
			Log(LogTypeStage, fmt.Sprintf("Stage %s PASSED", name), "")
		}
	}()
	fn()
	return nil
}
