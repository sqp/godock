///bin/true; exec /usr/bin/env go run "$0" "$@"

package main

import (
	"github.com/sqp/godock/libs/log" // Display info in terminal.
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/libs/packages/editpack"
)

func main() {
	logger := log.NewLog(log.Logs)

	externalUserDir, e := packages.DirAppletsExternal("") // option config dir
	log.Fatal(e, "DirAppletsExternal")

	packs, e := editpack.PacksExternal(logger, externalUserDir)
	log.Fatal(e, "get external packages")

	editpack.Start(logger, packs)
}
