package main

import (
	"flag"
	"fmt"
	"gitern/account"
	"gitern/art"
	"gitern/misc"
	"gitern/repo"
	"os"
	"os/exec"
)

func main() {
	var public bool
	flag.BoolVar(&public, "public", false, "make public")
	flag.Parse()

	if len(flag.Args()) == 0 {
		art.Fool.Fatal("Usage: gitern-create [--public] <path>")
	}

	free, err := misc.EnvToLimit("FREE")
	if err != nil {
		art.Tower.Fatal("free env", err.Error())
	}
	if free <= 0 {
		accname := os.Getenv("ACTIVE_ACCOUNT")
		fp := os.Getenv("FP")

		id, err := account.CreateSession(accname, fp)
		if err != nil {
			art.Tower.Fatal("create session", err.Error())
		}

		art.Scales.Fatal(
			fmt.Sprintf("Your account quota for '%s' of 25MB is full.", accname),
			fmt.Sprintf("Get unlimited storage on gitern; add a payment method to '%s.'", accname),
			fmt.Sprintf("Visit gitern.com/%s in a web browser.", id),
		)
	}

	path := repo.CanonicalizeRepoPath(flag.Arg(0))
	err = repo.MakePath(path, public)
	if err != nil {
		art.Fool.Fatal(path, err.Error())
	}

	err = os.Chdir(path)
	if err != nil {
		art.Tower.Fatal(path, "chdir", err.Error())
	}

	gitInit := exec.Command("git", "init", "--bare")
	if err := gitInit.Run(); err != nil {
		art.Tower.Fatal(path, "exec git init --bare", err.Error())
	}

	repoPath := "git@gitern.com:" + path
	art.Magic.Print(
		fmt.Sprintf("Initialized empty%s Git repository at %s", func() string {
			if public {
				return " public"
			}
			return ""
		}(), repoPath),
		"To add as remote:",
		fmt.Sprintf("\tgit remote add origin %s", repoPath),
		"To clone the empty repository:",
		fmt.Sprintf("\tgit clone %s", repoPath),
	)
}
