package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

func mountpoint(src, accountPath string) string {
	IGNORE := []string{
		"git-shell-commands",
		"accounts",
	}
	RENAME := map[string]string{
		"git-shell-commands-MEDSEC": "git-shell-commands",
	}

	basename := filepath.Base(src)
	for _, name := range IGNORE {
		if name == basename {
			return ""
		}
	}

	name, ok := RENAME[basename]
	if ok {
		basename = name
	}
	return filepath.Join(accountPath, basename)
}

// commissary is at the root of force-command's jail
func mountCommissary(accountPath string) error {
	files, err := ioutil.ReadDir("/")
	if err != nil {
		return err
	}

	for _, file := range files {
		src := filepath.Join("/", file.Name())
		dest := mountpoint(src, accountPath)
		if dest == "" {
			continue
		}

		err = createDir(dest, 0, 0)
		if err != nil {
			return err
		}
		err = mount(src, dest)
		if err != nil {
			return err
		}
	}
	return nil
}

// create directory with account name owned by git
// which will host the account's repos
// e.g. <ACCOUNTS_PATH>/<account name>/<account name>
// when user is chrooted in <ACCOUNTS_PATH>/<account name> their
// repos have the path <account name>/some/repo/path.git
func createRepoDir(accountPath string) error {
	gid := os.Getgid()
	dirPath := filepath.Join(accountPath, filepath.Base(accountPath))
	return createDir(dirPath, realUserId, gid)
}

func createCell(accountPath string) error {
	err := mountCommissary(accountPath)
	if err != nil {
		return err
	}

	err = createRepoDir(accountPath)
	if err != nil {
		return err
	}

	return nil
}

func createDir(path string, uid, gid int) error {
	return doPriveleged(func() error {
		err := os.Mkdir(path, 0755)
		if err != nil {
			if os.IsExist(err) {
				return nil
			}
			return err
		}

		if uid != 0 || gid != 0 {
			return os.Chown(path, uid, gid)
		}

		return nil
	})
}

func mount(src, dest string) error {
	files, err := ioutil.ReadDir(dest)
	if err != nil {
		return err
	}

	// we assume that all commissary dirs have contents, so
	// we use that to decide whether to mount them or not
	if len(files) == 0 {
		return doPriveleged(func() error {
			// mount readonly, don't allow set uid
			var opts uintptr = syscall.MS_BIND | syscall.MS_RDONLY | syscall.MS_NOSUID
			return syscall.Mount(src, dest, "none", opts, "")
		})
	}

	return nil
}
