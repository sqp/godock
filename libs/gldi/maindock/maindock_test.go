package maindock

import (
	"github.com/stretchr/testify/assert"
	"os"
	"os/user"
	"testing"
)

// func (dir string) string {

func TestConfigDir(t *testing.T) {
	usr, _ := user.Current()
	current, _ := os.Getwd()

	assert.Equal(t, "/test/root", ConfigDir("/test/root"), "test root dir (start with /)")
	assert.Equal(t, usr.HomeDir+"/test/home", ConfigDir("~/test/home"), "test home dir (start with ~)")
	assert.Equal(t, current+"/test/relative", ConfigDir("test/relative"), "test relative dir (not starting with /)")
	assert.Equal(t, usr.HomeDir+"/.config/"+CAIRO_DOCK_DATA_DIR, ConfigDir(""), "test empty dir (need default conf)")
}
