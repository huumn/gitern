package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const BASEPATH = "/repos"

// we check if parents are a .git dir
func insideRepo(path string) bool {
	dirs := strings.Split(filepath.Dir(path), "/")
	for _, d := range dirs {
		if strings.HasSuffix(d, ".git") {
			return true
		}
	}

	return false
}

func ValidPath(path string) bool {
	return path == filepath.Clean(path) &&
		!strings.HasPrefix(path, "..")
}

func ValidRepoPath(path string) bool {
	return ValidPath(path) && path != "/"
}

func CanonicalizePath(path string) string {
	return filepath.Join(BASEPATH, path)
}

func CanonicalizeRepoPath(path string) string {
	if !strings.HasSuffix(path, ".git") {
		path += ".git"
	}
	return path
}

// TODO: this is where we control whether a repo is public
// give all non-repo parents permissions of 754
// give the repo dir itself perms depending on if it's public
// if public 755, if not 750
func MakePath(path string, public bool) error {
	if insideRepo(path) {
		return fmt.Errorf("Inside a git repository")
	}

	if PathExists(path) {
		return fmt.Errorf("Already exists")
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	// if the repo is private we chmod the repo
	if !public {
		return os.Chmod(path, 0750)
	}

	return nil
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func PathPublic(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}

	// this is a public repo if it has "everyone" perms set
	return info.Mode().Perm()&0007 > 0
}

func CleanParents(path string) {
	// remove empty parents of path
	parent := filepath.Dir(path)
	for parent != BASEPATH {
		err := os.Remove(parent)
		// if this is not an empty dir we expect an err
		// and return. Because this is a nice to have
		// rather than a must have we ignore other errors
		if err != nil {
			break
		}
		parent = filepath.Dir(parent)
	}
}
