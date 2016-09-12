package packages_test

import (
	"github.com/stretchr/testify/assert"

	"github.com/sqp/godock/libs/log" // Display info in terminal.
	"github.com/sqp/godock/libs/packages"

	"os"
	"strings"
	"testing"
)

const appDir = "../../applets"

var logger = log.NewLog(log.Logs)

func TestListFromDir(t *testing.T) {
	packs, e := packages.ListFromDir(logger, appDir, packages.TypeGoInternal, packages.SourceApplet)
	if !assert.NoError(t, e, "ListFromDir") || !assert.NotEmpty(t, packs, "ListFromDir") {
		return
	}
	assert.NotEmpty(t, packs[0].DisplayedName, "DisplayedName")
}

func TestNewPack(t *testing.T) {
	appname := "Audio"

	// Load package.
	pack, e := packages.NewAppletPackageUser(logger, appDir, appname, packages.TypeGoInternal, packages.SourceApplet)
	assert.NoError(t, e, "NewAppletPackageUser")
	validatePack(t, pack)

	// Read values.
	assert.Equal(t, appname, pack.DisplayedName, "DisplayedName")
	assert.Equal(t, "SQP", pack.Author, "Author")
	assert.Equal(t, "0.0.4", pack.Version, "Version")
	assert.Equal(t, 6, pack.Category, "Category")
	assert.False(t, pack.IsMultiInstance, "IsMultiInstance")
	assert.True(t, pack.ActAsLauncher, "ActAsLauncher")
	assert.True(t, strings.HasPrefix(pack.Description, "Pulseaudio"), "Description")
}

func validatePack(t *testing.T, pack *packages.AppletPackage) {
	if !assert.NotNil(t, pack, "NewAppletPackageUser") {
		assert.Fail(t, "package nil")
	}
}

func TestDir(t *testing.T) {
	home := os.Getenv("HOME")
	for in, out := range map[string]string{
		"":     home + "/.config/cairo-dock/third-party",
		"/tmp": "/tmp/third-party",
	} {
		str, e := packages.DirAppletsExternal(in)
		assert.NoError(t, e, "DirAppletsExternal")
		assert.Equal(t, out, str, "DirAppletsExternal")
	}

	for in, out := range map[string]string{
		"":     home + "/.config/cairo-dock/extras/gauges",
		"/tmp": "/tmp/extras/gauges",
	} {
		str, e := packages.DirThemeExtra(in, "gauges")
		assert.NoError(t, e, "DirTheme")
		assert.Equal(t, out, str, "DirTheme")
	}

	gauges, e := packages.ListThemesDir(logger, "/usr/share/cairo-dock/gauges", packages.TypeLocal)
	assert.NoError(t, e, "ListThemesDir")
	assert.True(t, len(gauges) > 3, "ListThemesDir count")
	gauge := gauges[0]
	assert.Equal(t, "Battery[0]", gauge.GetName(), "gauge.GetName()")
	assert.Equal(t, "Battery", gauge.GetTitle(), "gauge.GetTitle()")
	assert.Equal(t, "Adrien Pilleboue", gauge.GetAuthor(), "gauge.GetAuthor()")
	assert.Equal(t, "2", gauge.GetModuleVersion(), "gauge.GetModuleVersion()")
	assert.Equal(t, packages.TypeLocal, gauge.Type, "gauge.Type")
	// assert.Equal(t, "", gauge.GetGettextDomain(), "gauge.GetGettextDomain()")
	assert.Equal(t, "Made by Necropotame for Cairo-Dock.\n", gauge.GetDescription(), "gauge.GetDescription()")
	assert.Equal(t, "/usr/share/cairo-dock/gauges/Battery/preview", gauge.GetPreviewFilePath(), "gauge.GetPreviewFilePath()")
}
