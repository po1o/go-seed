package main

import "testing"

// TestMainDoesNotPanic is a smoke test ensuring the entrypoint wiring stays intact.
func TestMainDoesNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("main panicked: %v", r)
		}
	}()

	main()
}
