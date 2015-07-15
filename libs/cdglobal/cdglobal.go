// Package cdglobal defines application and backend global consts and data.
package cdglobal

import "os"

// AppVersion defines the application version.
//
// The -git suffix is used to tag the default value, but it should be removed if
// the Makefile was used for the build, and the real version was set.
//
var AppVersion = "0.0.3.2-git"

// Variables set at build time.
var (
	GitHash   = ""
	BuildDate = ""

	AppBuildPath = []string{"github.com", "sqp", "godock"}
)

// AppBuildPathFull returns a splitted path to the application build directory.
//
func AppBuildPathFull() []string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return nil
	}

	return append([]string{gopath, "src"}, AppBuildPath...)
}
