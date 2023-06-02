package main

import (
	"gitern/art"
	"gitern/repo"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		art.Fool.Fatal("Usage: gitern-delete <path>")
	}

	path := repo.CanonicalizeRepoPath(os.Args[1])

	if !repo.PathExists(path) {
		art.Fool.Fatal(path, "Does not exist")
	}

	err := os.RemoveAll(path)
	if err != nil {
		art.Tower.Fatal(path, "remove", err.Error())
	}

	repo.CleanParents(path)

	art.Skull.Print(path, "deleted")
}
