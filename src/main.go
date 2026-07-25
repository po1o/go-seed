// Command go-seed is a minimal, batteries-included seed for new Go projects.
//
// It ships the tooling scaffolding (go tool linters, go-task tasks, CI
// workflows) so a new project starts from a known-good, lint-clean baseline.
package main

import (
	"fmt"
	"os"

	"github.com/po1o/go-seed/src/build"
)

func main() {
	if _, err := fmt.Fprintln(os.Stdout, "Hello, World!"); err != nil {
		os.Exit(1)
	}
	if _, err := fmt.Fprintf(os.Stdout, "go-seed %s\n", build.Version); err != nil {
		os.Exit(1)
	}
}
