package v1

import "fmt"

// TestError represents a controlled test failure.
type TestError struct {
	Message string
}

func (e TestError) Error() string {
	return e.Message
}

// Fail fails the current test stage with a message.
// It uses panic with TestError to stop execution, which is caught by the Stage runner.
func Fail(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	// In dry-run mode we only log and skip panicking so that action discovery can proceed
	// without executing real operations or being stopped by setup checks.
	if IsDryRun() {
		Log(LogTypeError, "Assertion FAILED (dry-run, no panic)", msg)
		return
	}

	Log(LogTypeError, "Assertion FAILED", msg)
	panic(TestError{Message: msg})
}

// Assert checks if the condition is true. If not, it fails the test stage.
func Assert(condition bool, format string, args ...interface{}) {
	if !condition {
		Fail(format, args...)
	}
}

// AssertNoError asserts that the error is nil.
func AssertNoError(err error) {
	if err != nil {
		Fail("Unexpected error: %v", err)
	}
}
