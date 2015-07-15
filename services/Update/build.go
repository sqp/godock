package Update

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/packages/build"
	"github.com/sqp/godock/libs/packages/versions"

	"path/filepath"
	"text/template"
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
		name := app.conf.BuildTargets[app.targetID]
		app.target = build.NewBuilder(name, app.Log())
		app.target.SetProgress(func(f float64) { app.RenderValues(f) })
		app.Log().Debug("setBuildTarget", app.target.Label())

		switch target := app.target.(type) {

		case *build.BuilderCore:
			target.SetDir(filepath.Join(app.conf.SourceDir, app.conf.DirCore))

		case *build.BuilderApplets:
			target.SetDir(filepath.Join(app.conf.SourceDir, app.conf.DirApplets))
			target.MakeFlags = app.conf.FlagsApplets

		case *build.BuilderInternal:
			target.SetDir(filepath.Join(app.conf.SourceDir, app.conf.DirApplets))
		}
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
	switch build.GetSourceType(target) {
	case build.TypeAppletScript, build.TypeAppletCompiled:
		if target == app.Name() { // Don't eat the chicken, or you won't have any more eggs.
			app.Log().ExecAsync("make", "reload")

		} else {
			build.AppletRestart(target)
		}

	case build.TypeGodock:

	default:
		func() {
			app.Log().ExecAsync("cdc", "restart")
		}()

	}
}

//
//---------------------------------------------------------[ VERSION POLLING ]--

// Versions handles version checking on multiple VCS repositories.
//
type Versions struct {
	sources        []*Repo
	template       *template.Template
	dialogTemplate string
	fileTemplate   string
	newCommits     int

	// Polling data.
	restart    chan bool        // restart channel to forward user requests.
	callResult func(int, error) // Action to execute to send polling results.
}

// Sources lists managed repositories.
//
func (ver *Versions) Sources() []*Repo {
	return ver.sources
}

// Check is the callback for poller.
// TODO: need to better handle errors.
//
func (ver *Versions) Check() {
	var nb, cur int
	var e error
	for _, repo := range ver.sources {
		cur, e = repo.findNew()
		nb += cur
	}
	ver.callResult(nb, e)
}

//
//-----------------------------------------------[ SOURCES BRANCH MANAGEMENT ]--

// Repo checks and updates a VCS sources repository.
//
type Repo struct {
	Name    string // Repository name.
	Dir     string // Location of repo on the filesystem.
	GotData bool   // true if data was successfully pulled.

	Log   string // Commit messages for new commits.
	Delta int    // Delta of revisions between server and local.

	// template display fields.
	NewLocal int  //  Number of unmerged patch in local dir.
	Zero     bool // True if revisions are the same.

	logger cdtype.Logger
}

// NewRepo creates a source repo with name and dir.
//
func NewRepo(log cdtype.Logger, repo, dir string) *Repo {
	return &Repo{
		Name:   repo,
		Dir:    dir,
		logger: log,
	}
}

// findNew gets revisions informations.
//
func (repo *Repo) findNew() (new int, e error) {
	repo.logger.Debug("Get version", repo.Name)
	repo.GotData = false
	v, e := versions.NewFetched(repo.Dir)
	if e != nil {
		return 0, e
	}

	// get the number of new commits on the server.
	count, e := v.CountCommits("HEAD...origin/master")
	if e != nil {
		return 0, e
	}

	// We have valid data.

	repo.GotData = true
	repo.Delta = count

	repo.Log, e = v.DeltaLog("HEAD...origin/master")
	repo.logger.Err(e, "log", repo.Name)

	switch { // Data for formatter.
	case repo.Delta == 0:
		repo.Zero = true
	case repo.Delta < 0:
		repo.NewLocal = -repo.Delta
	}

	return repo.Delta, e
}

// update updates the local repo to server version, returns the count and log.
//
func (repo *Repo) update() (delta int, logstr string, e error) { // , progress func(float64)
	v, e := versions.NewFetched(repo.Dir)
	if e != nil {
		return 0, "", e
	}

	repo.logger.Info("download", repo.Name, ":", repo.Dir)
	rev, _ := v.Rev("HEAD")
	repo.logger.Info("current:", rev)

	logstr, e = v.DeltaLog("origin..." + rev)
	if e != nil {
		return 0, "", e
	}

	delta, e = v.Update()
	if e != nil {
		return 0, "", e
	}
	rev, _ = v.Rev("HEAD")
	repo.logger.Info("new rev:", rev)
	if delta > 0 {
		repo.logger.Info("imported", delta, "commit(s)")
		println(logstr)
	} else {
		repo.logger.Info("no change")
	}
	return delta, logstr, e
}
