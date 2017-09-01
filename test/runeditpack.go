///bin/true; exec /usr/bin/env go run "$0" "$@"

package main

import (
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/log" // Display info in terminal.
	"github.com/sqp/godock/libs/packages/editpack"
)

func main() {
	logger := log.NewLog(log.Logs)

	externalUserDir, e := cdglobal.DirAppletsExternal("") // option config dir
	if logger.Err(e, "DirAppletsExternal") {
		return
	}

	packs, e := editpack.PacksExternal(logger, externalUserDir)
	if logger.Err(e, "get external packages") {
		return
	}

	editpack.Start(logger, packs)
}
