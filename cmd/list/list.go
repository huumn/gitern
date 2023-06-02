package main

import (
	"bufio"
	"fmt"
	"gitern/account"
	"gitern/art"
	"gitern/repo"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func printWalk(w io.Writer, path string, rel bool, privateAccess bool) {
	_, err := os.Stat(path)
	if err != nil {
		return
	}

	err = filepath.Walk(path, func(fpath string, info os.FileInfo, err error) error {
		// we ignore the errors as they are likely perm error when we don't have private access
		if err != nil {
			// we shouldn't have errors if we have private access
			if privateAccess {
				art.Tower.Fatal(path, "inner walk", err.Error())
			}

			// if we don't have private access, it's none of their business
			// we also expect to have perms errors
			return nil
		}
		if info.IsDir() && len(info.Name()) > 4 &&
			strings.HasSuffix(info.Name(), ".git") {
			if rel {
				fpath, err = filepath.Rel(path, fpath)
				if err != nil {
					art.Tower.Fatal(path, "make relative", err.Error())
				}
			}

			if repo.PathPublic(fpath) {
				fmt.Fprintln(w, fpath+" (public)")
			} else if privateAccess {
				fmt.Fprintln(w, fpath)
			}
		}
		return nil
	})
	if err != nil {
		art.Tower.Fatal(path, "outer walk", err.Error())
	}
}

func main() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	if len(os.Args) == 1 {
		accounts, err := account.Accounts()
		if err != nil {
			art.Tower.Fatal("get accounts", err.Error())
		}
		for _, acc := range accounts {
			path, err := acc.Path()
			if err != nil {
				art.Tower.Fatal("account path", err.Error())
			}

			err = os.Chdir(path)
			if err != nil {
				art.Tower.Fatal("account path", err.Error())
			}

			printWalk(w, acc.Name, false, true)
		}
	} else if len(os.Args) == 2 {
		accName := os.Args[1]
		printWalk(w, accName, false, account.CheckAccess(accName) == nil)
	} else {
		art.Fool.Fatal("Usage: gitern-list [<path>]")
	}
}
