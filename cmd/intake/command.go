package main

import (
	"os"
	"strings"
	"syscall"
)

type command struct {
	name string
	args []string
	full string
}

func newCommand(cmdStr string) command {
	// log.Printf("CMD %s UID %d EUID %d GID %d EGID %d\n", cmdStr, syscall.Getuid(),
	// 	syscall.Geteuid(), syscall.Getgid(), syscall.Getegid())

	fields := strings.Fields(cmdStr)
	if len(fields) > 0 {
		return command{fields[0], fields[1:], cmdStr}
	}
	return command{}
}

func (c command) serfsAllowed() bool {
	SERF_CMDS := map[string]int{
		"gitern-list":        1,
		"git-upload-pack":    5,
		"git-upload-archive": 1,
	}

	nargs, ok := SERF_CMDS[c.name]
	return ok && len(c.args) <= nargs
}

func (c command) noAccountArgNeeded() bool {
	MIN_SEC_CMDS := map[string]int{
		"gitern-pubkey-add":    1,
		"gitern-pubkey-remove": 2,
		"gitern-pubkey-list":   2,
		"gitern-list":          0,
		"":                     0,
	}

	nargs, ok := MIN_SEC_CMDS[c.name]
	return ok && len(c.args) <= nargs
}

func (c command) allowedInMinSecurity() bool {
	MIN_SEC_CMDS := map[string]int{
		"gitern-pubkey-add":    1,
		"gitern-pubkey-remove": 2,
		"gitern-pubkey-list":   2,
		"gitern-list":          0,
		"":                     0,
	}

	nargs, ok := MIN_SEC_CMDS[c.name]
	return ok && len(c.args) <= nargs
}

func (c command) accountViaPath() bool {
	return c.name == "gitern-create" ||
		c.name == "gitern-delete" ||
		c.name == "gitern-list" ||
		c.name == "git-receive-pack" ||
		c.name == "git-upload-pack" ||
		c.name == "git-upload-archive"
}

// we assume the last arg (if there is one) is the path
func (c command) accountTarget() string {
	trimmedPath := strings.TrimPrefix(strings.TrimPrefix(c.path(), "'"),
		string(os.PathSeparator))
	return strings.Split(trimmedPath, string(os.PathSeparator))[0]
}

func (c command) path() string {
	if len(c.args) < 1 {
		return ""
	}
	return c.args[len(c.args)-1]
}

func (c command) exec() error {
	// log.Printf("EXEC as UID %d EUID %d GID %d EGID %d\n", syscall.Getuid(),
	// 	syscall.Geteuid(), syscall.Getgid(), syscall.Getegid())

	// exec SSH_ORIGINAL_COMMAND
	argv := []string{"/usr/bin/git-shell"}
	if c.full != "" {
		argv = append(argv, "-c", c.full)
	}
	err := syscall.Exec(argv[0], argv, os.Environ())
	if err != nil {
		return err
	}

	return nil
}
