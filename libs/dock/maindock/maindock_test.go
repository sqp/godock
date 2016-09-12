package maindock

import (
	"github.com/sqp/godock/libs/cdglobal"
	liblog "github.com/sqp/godock/libs/log" // Display info in terminal.

	"fmt"
	"testing"
)

func TestHidden(t *testing.T) {
	SetLogger(liblog.NewLog(liblog.Logs))
	altConfigPath := ""
	confdir := cdglobal.ConfigDirDock(altConfigPath)
	hidden := loadHidden(confdir)
	fmt.Println(hidden.DefaultBackend)
	fmt.Println(hidden.LastVersion)
	fmt.Println(hidden.LastYear)
	fmt.Println(hidden.SessionWasUsed)
}
