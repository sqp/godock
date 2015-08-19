package versions

import (
	"github.com/sqp/godock/libs/cdtype"

	"text/template"
)

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

func NewVersions(callResult func(int, error), repos ...*Repo) *Versions {
	return &Versions{
		callResult: callResult,
		newCommits: -1,
		sources:    repos,
	}
}

// CountNew returns the number of new commits.
//
func (ver *Versions) CountNew() int {
	return ver.newCommits
}

// SetCountNew sets the number of new commits.
//
func (ver *Versions) SetCountNew(countNew int) {
	ver.newCommits = countNew
}

// Sources lists managed repositories.
//
func (ver *Versions) Sources() []*Repo {
	return ver.sources
}

// AddSources adds repositories to the managed list.
//
func (ver *Versions) AddSources(repos ...*Repo) *Versions {
	ver.sources = append(ver.sources, repos...)
	return ver
}

// Clear clears the repositories list.
//
func (ver *Versions) Clear(repos ...*Repo) *Versions {
	ver.sources = nil
	return ver
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
	v, e := NewFetched(repo.Dir)
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

// Update updates the local repo to server version, returns the count and log.
//
func (repo *Repo) Update() (delta int, logstr string, e error) { // , progress func(float64)
	v, e := NewFetched(repo.Dir)
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
