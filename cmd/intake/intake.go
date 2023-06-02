package main

import (
	"gitern/account"
	"gitern/art"
	"gitern/misc"
	"os"
	"strconv"
	"syscall"
)

// force command
// 1. mounts all deps in the account dir ACCOUNT
// 2. chroots us to the account dir ACCOUNT (if required)
// 3. execs SSH_ORIGINAL_COMMAND

// https://www.gnu.org/software/libc/manual/html_node/Setuid-Program-Example.html
// https://www.oreilly.com/library/view/secure-programming-cookbook/0596003943/ch01s03.html
var (
	realUserId      int
	effectiveUserId int
)

const (
	SerfUID = 444
	SerfGID = 445
)

func unprivelege() {
	err := syscall.Setreuid(effectiveUserId, realUserId)
	if err != nil {
		art.Tower.Fatal("unprivelege", err.Error())
	}
}

func privelege() {
	err := syscall.Setreuid(realUserId, effectiveUserId)
	if err != nil {
		art.Tower.Fatal("privelege", err.Error())
	}
}

func bondSerf() {
	privelege()

	err := syscall.Setgroups([]int{SerfGID}) // remove sup group priveleges too
	if err != nil {
		art.Tower.Fatal("bonding serf (sup groups)", err.Error())
	}

	// must set group before user id
	err = syscall.Setgid(SerfGID)
	if err != nil {
		art.Tower.Fatal("bonding serf (gid)", err.Error())
	}

	err = syscall.Setuid(SerfUID)
	if err != nil {
		art.Tower.Fatal("bonding serf (uid)", err.Error())
	}
}

func doPriveleged(fn func() error) error {
	privelege()
	err := fn()
	unprivelege()
	return err
}

func lockup(accPath string) error {
	err := doPriveleged(func() error {
		return syscall.Chroot(accPath)
	})
	if err != nil {
		return err
	}

	return os.Chdir("/")
}

func main() {
	realUserId = syscall.Getuid()
	effectiveUserId = syscall.Geteuid()
	syscall.Setgroups([]int{syscall.Getgid()}) // remove sup group priveleges too
	unprivelege()

	command := newCommand(os.Getenv("SSH_ORIGINAL_COMMAND"))

	if command.allowedInMinSecurity() {
		err := command.exec()
		art.Tower.Fatal("exec", err.Error()) // if we get here, error
	}

	accName := command.accountTarget()
	if accName == "" {
		if command.accountViaPath() {
			art.Fool.Fatal("you must specify a repo path")
		}

		art.Fool.Fatal("you must specify an account name")
	}

	accPath, err := account.AccountPath(accName)
	if err != nil {
		art.Fool.Fatal("no such account", accName)
	}

	err = createCell(accPath)
	if err != nil {
		art.Tower.Fatal("create cell", err.Error())
	}

	err = lockup(accPath)
	if err != nil {
		art.Tower.Fatal("lockup", err.Error())
	}

	err = account.CheckAccess(accName)
	if err == nil {
		authAcc, err := account.GetAuthAccount(accName)
		if err != nil {
			art.Tower.Fatal("could not get lord's account", err.Error())
		}

		if authAcc.Quota != 0 {
			repoPath := authAcc.ReposPath()
			du, err := misc.DiskUsage(repoPath)
			if err != nil {
				art.Tower.Fatal("disk usage", err.Error())
			}
			os.Setenv("QUOTA", strconv.FormatInt(authAcc.Quota, 10))
			os.Setenv("FREE", strconv.FormatInt(authAcc.Quota-du, 10))
		}
		os.Setenv("ACTIVE_ACCOUNT", authAcc.Name)
	} else {
		// are we allowed to execute the command?
		if !command.serfsAllowed() {
			art.Fool.Fatal(err.Error())
		}

		// give them the lowest perms
		bondSerf()
	}

	err = command.exec()
	art.Tower.Fatal("exec", err.Error()) // if we get here, error
}
