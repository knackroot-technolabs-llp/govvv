package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type git struct {
	dir string
}

func (g git) exec(args ...string) (string, error) {
	var errOut bytes.Buffer
	c := exec.Command("git", args...)
	c.Dir = g.dir
	c.Stderr = &errOut
	out, err := c.Output()
	outStr := strings.TrimSpace(string(out))
	if err != nil {
		err = fmt.Errorf("git: error=%q stderr=%s", err, string(errOut.Bytes()))
	}
	return outStr, err
}

// Commit returns the short git commit hash.
func (g git) Commit() (string, error) {
	return g.exec("rev-parse", "--short", "HEAD")
}

// isStateDirty returns true if the repository's state is diry
func (g git) isStateDirty() (bool, error) {
	out, err := g.exec("status", "--porcelain")
	if err != nil {
		return false, err
	}
	if len(out) > 0 {
		return true, nil
	}

	return false, nil
}

// State returns the repository state indicating whether
// it is "clean" or "dirty".
func (g git) State() (string, error) {
	isDirty, err := g.isStateDirty()
	if err != nil {
		return "", err
	}
	if isDirty {
		return "dirty", nil
	}
	return "clean", nil
}

// Branch returns the branch name. If it is detached,
// or an error occurs, returns "HEAD".
func (g git) Branch() string {
	out, err := g.exec("symbolic-ref", "-q", "--short", "HEAD")
	if err != nil {
		// might be failed due to another reason, but assume it's
		// exit code 1 from `git symbolic-ref` in detached state.
		return "HEAD"
	}
	return out
}

// Summary returns the tag that the current commit has, or the commit hash otherwise.
// If the working directory is not clean, it appends the return value with "-dirty"
func (g git) Summary() (string, error) {
	// see if we have a tag to use as the prefix
	id, err := g.exec("describe", "--tags", "--exact-match", "--always")
	if err != nil || id == "" {
		// we'll assume the error is because there is no tag on this commit
		// let's now try to use the commit hash as the id
		id, err = g.Commit()
		if err != nil {
			return "", err
		}
	}

	dirty, err := g.isStateDirty()
	if err != nil || !dirty { // in case of error, we'll not declare it dirty
		return id, nil
	}

	return id + "-dirty", nil
}
