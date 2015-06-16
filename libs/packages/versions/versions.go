// Package versions checks and updates vcs packages. Only git supported.
package versions

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// Git commands args.
var (
	ArgsFetch        = "fetch"
	ArgsRev          = "rev-parse {location}"        // HEAD  or origin
	ArgsCountCommits = "rev-list --count {location}" //
	ArgsDeltaLog     = "log {format} {location}"     // -n {limit}
	ArgsUpdate       = "pull --ff-only"
)

var (
	// LogFormat defines the format for ShowVersions: just show commits title (one line).
	LogFormat = "--pretty=format:%s"

// LogFormat = "--oneline"
// LogFormat = "--pretty=format:\"%an : %s\""
)

// VCS get informations and acts on a git repository.
//
type VCS struct {
	Cmd string
	Dir string
}

// New creates a git client for the repo.
//
func New(repoRoot string) (*VCS, error) {

	// Try and find the root of the VCS repo using repoRoot as a starting point.
	if repoRoot == "" {
		return nil, errors.New("repoRoot empty")
	}

	vcsRoot, e := fromDir(repoRoot)
	if e != nil {
		return nil, e
	}

	if vcsRoot == "" {
		return nil, errors.New("fucked")
	}
	return &VCS{
		Cmd: "git",
		Dir: vcsRoot,
	}, nil
}

// NewFetched creates a git client for the repo and updates server info.
//
func NewFetched(repoRoot string) (*VCS, error) {
	v, e := New(repoRoot)
	if e != nil {
		return nil, e
	}
	e = v.Fetch() // Download server informations (apt-get update).
	if e != nil {
		return nil, e
	}
	return v, nil
}

// Fetch downloads server informations (apt-get update).
//
func (v *VCS) Fetch() error {
	_, e := v.RunOutput(v.Dir, ArgsFetch)
	return e
}

// Rev returns the current commit ID for location.
//
func (v *VCS) Rev(location string) (rev string, e error) {
	b, e := v.RunOutput(v.Dir, ArgsRev, "location", location)
	return strings.TrimSuffix(string(b), "\n"), e
}

// CountCommits gives the total number of commits of the branch at location.
//
func (v *VCS) CountCommits(location string) (int, error) {
	l, er := v.RunOutput(v.Dir, ArgsCountCommits, "location", location)
	if er != nil {
		return 0, er
	}

	str := strings.TrimSuffix(string(l), "\n")
	delta, ec := strconv.Atoi(str)
	return delta, ec
}

// DeltaLog shows logs for location (or query).
//
func (v *VCS) DeltaLog(location string) (string, error) {
	l, e2 := v.RunOutput(v.Dir, ArgsDeltaLog, "format", LogFormat, "location", location)
	return string(l), e2
}

// Update updates the local branch and returns the number of new commmits.
//
func (v *VCS) Update() (new int, e error) {
	before, _ := v.CountCommits("HEAD")

	_, e = v.RunOutput(v.Dir, ArgsUpdate)
	if e != nil {
		return 0, e
	}

	after, _ := v.CountCommits("HEAD")

	return after - before, e
}

//

// fromDir inspects dir and its parents to determine the
// version control system and code repository to use.
//
func fromDir(dir string) (path string, err error) {
	dir = filepath.Clean(dir)

	for len(dir) > 1 {
		if fi, err := os.Stat(filepath.Join(dir, ".git")); err == nil && fi.IsDir() {
			return dir, nil
		}

		// Move to parent.
		ndir := filepath.Dir(dir)
		if len(ndir) >= len(dir) {
			// Shouldn't happen, but just in case, stop.
			break
		}
		dir = ndir
	}

	return "", fmt.Errorf("directory %q is not using a known version control system", dir)
}

//

//
//-------------------------------------------------------------[ FROM VCS.GO ]--

// from code.google.com/p/go.tools/go/vcs/vcs.go

// updated by SQP to work as standalone.

// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style

// RunOutput is like run but returns the output of the command.
func (v *VCS) RunOutput(dir string, cmd string, keyval ...string) ([]byte, error) {
	return v.run1(dir, cmd, keyval, true)
}

// run1 is the generalized implementation of run and runOutput.
func (v *VCS) run1(dir string, cmdline string, keyval []string, verbose bool) ([]byte, error) {
	m := make(map[string]string)
	for i := 0; i < len(keyval); i += 2 {
		m[keyval[i]] = keyval[i+1]
	}
	args := strings.Fields(cmdline)
	for i, arg := range args {
		args[i] = expand(m, arg)
	}

	_, err := exec.LookPath(v.Cmd)
	if err != nil {
		// fmt.Fprintf(os.Stderr,
		// 	"go: missing %s command. See http://golang.org/s/gogetcmd\n",
		// 	v.Name)
		return nil, err
	}

	cmd := exec.Command(v.Cmd, args...)
	cmd.Dir = dir
	cmd.Env = envForDir(cmd.Dir)
	// if ShowCmd {
	// 	fmt.Printf("cd %s\n", dir)
	// 	fmt.Printf("%s %s\n", v.Cmd, strings.Join(args, " "))
	// }
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	out := buf.Bytes()
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "# cd %s; %s %s\n", dir, v.Cmd, strings.Join(args, " "))
			os.Stderr.Write(out)
		}
		return nil, err
	}
	return out, nil
}

// expand rewrites s to replace {k} with match[k] for each key k in match.
func expand(match map[string]string, s string) string {
	for k, v := range match {
		s = strings.Replace(s, "{"+k+"}", v, -1)
	}
	return s
}

//

// from code.google.com/p/go.tools/go/vcs/env.go

// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style

// envForDir returns a copy of the environment
// suitable for running in the given directory.
// The environment is the current process's environment
// but with an updated $PWD, so that an os.Getwd in the
// child will be faster.
func envForDir(dir string) []string {
	env := os.Environ()
	// Internally we only use rooted paths, so dir is rooted.
	// Even if dir is not rooted, no harm done.
	return mergeEnvLists([]string{"PWD=" + dir}, env)
}

// mergeEnvLists merges the two environment lists such that
// variables with the same name in "in" replace those in "out".
func mergeEnvLists(in, out []string) []string {
NextVar:
	for _, inkv := range in {
		k := strings.SplitAfterN(inkv, "=", 2)[0]
		for i, outkv := range out {
			if strings.HasPrefix(outkv, k) {
				out[i] = inkv
				continue NextVar
			}
		}
		out = append(out, inkv)
	}
	return out
}
