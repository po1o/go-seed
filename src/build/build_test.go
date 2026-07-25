package build_test

import (
	"testing"

	"github.com/po1o/go-seed/src/build"
)

func TestVersionHasDefault(t *testing.T) {
	t.Parallel()

	if build.Version == "" {
		t.Fatal("build.Version must not be empty")
	}
}
