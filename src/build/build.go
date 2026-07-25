// Package build carries build-time metadata for the go-seed binary.
package build

// Version is the go-seed version. It defaults to a development marker and is
// overridden at release time via -ldflags "-X .../build.Version=x.y.z".
var Version = "0.0.0-dev"
