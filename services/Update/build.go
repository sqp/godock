package Update

import (
	"github.com/sqp/godock/libs/packages/build"
	"github.com/sqp/godock/libs/srvdbus/dlogbus"

	"path/filepath"
	"time"
)

//
//----------------------------------------------------------[ SOURCES BUILDER ]--

// setBuildTarget creates the target renderer/builder.
//
func (app *Applet) setBuildTarget() {
	if !app.conf.UserMode { // Tester mode.
		app.target = &build.BuilderNull{}

	} else {
		var (
			name       = app.conf.BuildTargets[app.targetID]
			sourceType = build.GetSourceType(name, app.Log())
		)
		app.target = app.newBuilder(sourceType, name)
		app.Log().Debug("setBuildTarget", app.target.Label())
	}

	// Delayed display of emblem. 5ms seemed to be enough but 500 should do the job.
	go func() { time.Sleep(500 * time.Millisecond); app.showTarget() }()
}

// showTarget shows build target informations on the icon (label and emblem).
//
func (app *Applet) showTarget() {
	app.SetEmblem(app.target.Icon(), EmblemTarget)
	app.SetLabel("Target: " + app.target.Label())
}

// restartTarget restarts the current build target if needed,
// with everything necessary (like the dock for an internal app).
//
func (app *Applet) restartTarget() {
	if !app.conf.BuildReload {
		return
	}
	target := app.conf.BuildTargets[app.targetID]
	app.Log().Info("restart", target)
	switch build.GetSourceType(target, app.Log()) {
	case build.TypeAppletScript, build.TypeAppletCompiled:
		if target == app.Name() { // Don't eat the chicken, or you won't have any more eggs.
			app.Log().ExecAsync("make", "reload")

		} else {
			build.AppletRestart(target)
		}

	case build.TypeGodock:
		build.CloseGui()
		go app.Log().Err(dlogbus.Action((*dlogbus.Client).Restart), "restart") // No need to wait an answer, it blocks.

	default:
		func() {
			app.Log().ExecAsync("cdc", "restart")
		}()

	}
}

// newBuilder creates a builder of the given type with all settings.
// The name is mandatory for a single applet (internal or not).
func (app *Applet) newBuilder(sourceType build.SourceType, name string) build.Builder {
	bt := build.NewBuilder(sourceType, name, app.Log())
	bt.SetProgress(func(f float64) { app.DataRenderer().Render(f) })
	// bt.SetLogger(app.Log())

	switch target := bt.(type) {

	case *build.BuilderCore:
		target.SetDir(filepath.Join(app.conf.SourceDir, app.conf.DirCore))

	case *build.BuilderApplets:
		target.SetDir(filepath.Join(app.conf.SourceDir, app.conf.DirApplets))
		target.MakeFlags = app.conf.FlagsApplets

	case *build.BuilderInternal:
		target.SetDir(filepath.Join(app.conf.SourceDir, app.conf.DirApplets))
	}
	return bt
}
