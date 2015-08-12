package cdglobal

import (
	"github.com/stretchr/testify/assert"
	"os"
	"os/user"
	"testing"
)

func TestConfigDir(t *testing.T) {
	usr, _ := user.Current()
	current, _ := os.Getwd()

	for _, v := range []struct{ in, out, msg string }{
		{
			"/test/root",
			ConfigDirDock("/test/root"),
			"test root dir (start with /)",
		}, {
			usr.HomeDir + "/test/home",
			ConfigDirDock("~/test/home"),
			"test home dir (start with ~)",
		}, {
			current + "/test/relative",
			ConfigDirDock("test/relative"),
			"test relative dir (not starting with /)",
		}, {
			usr.HomeDir + "/.config/" + ConfigDirBaseName,
			ConfigDirDock(""),
			"test empty dir (need default conf)",
		},
	} {
		assert.Equal(t, v.in, v.out, v.msg)
	}
}
