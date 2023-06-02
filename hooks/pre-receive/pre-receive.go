package main

import (
	"bufio"
	"fmt"
	"gitern/account"
	"gitern/art"
	_ "gitern/logmill"
	"gitern/misc"
	"log"
	"os"
	"os/exec"

	"github.com/dustin/go-humanize"
)

// grab quota
// total up the size of the incoming if it causes the user to exceed their
// quota, warn

const NULLSHA = "0000000000000000000000000000000000000000"

func main() {
	free, err := misc.EnvToLimit("FREE")
	if err != nil {
		log.Fatalln(err)
	}
	scanner := bufio.NewScanner(os.Stdin)
	var oldref, newref string
	var totalSize int64
	for scanner.Scan() {
		_, err := fmt.Sscan(scanner.Text(), &oldref, &newref)
		if err != nil {
			log.Fatalln("scanning input", err)
		}

		if newref == NULLSHA {
			continue
		}

		var target string
		if oldref == NULLSHA {
			target = newref
		} else {
			target = fmt.Sprintf("%s..%s", oldref, newref)
		}

		// get all the incoming objects
		revList := exec.Command("git", "rev-list", "--objects", "--use-bitmap-index", target, "--not", "--branches=/*", "--tags=/*")
		// cat them out with their sizes
		catFile := exec.Command("git", "cat-file", "--buffer", "--batch-check=%(objectsize:disk) %(rest)")
		// pipe rev-list into cat-file
		catFile.Stdin, err = revList.StdoutPipe()
		if err != nil {
			log.Fatalln(err)
		}

		// capture cat-file's output
		catFilePipe, err := catFile.StdoutPipe()
		if err != nil {
			log.Fatalln(err)
		}

		err = catFile.Start()
		if err != nil {
			log.Fatalln(err)
		}

		err = revList.Start()
		if err != nil {
			log.Fatalln(err)
		}

		var size int64
		catFileScanner := bufio.NewScanner(catFilePipe)
		for catFileScanner.Scan() {
			_, err := fmt.Sscan(catFileScanner.Text(), &size)
			if err != nil {
				log.Fatalln("scanning catfile", err)
			}
			totalSize += size
		}

		if err := catFileScanner.Err(); err != nil {
			log.Fatalln(err)
		}

		err = catFile.Wait()
		if err != nil {
			log.Fatalln(err)
		}

		err = revList.Wait()
		if err != nil {
			log.Fatalln(err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}

	accname := os.Getenv("ACTIVE_ACCOUNT")
	fp := os.Getenv("FP")
	if totalSize > free {
		if free < 0 {
			free = 0
		}

		id, err := account.CreateSession(accname, fp)
		if err != nil {
			log.Fatalln(err)
		}

		art.Scales.Fatal(
			fmt.Sprintf("Commit of %s exceeds your remaining quota of %s on '%s.'",
				humanize.Bytes(uint64(totalSize)), humanize.Bytes(uint64(free)), accname),
			fmt.Sprintf("Get unlimited storage on gitern; add a payment method to '%s.'", accname),
			fmt.Sprintf("Visit gitern.com/%s in a web browser.", id),
		)
	} else if float64(free-totalSize) <= float64(totalSize)*0.2 {
		art.Scales.Print(
			fmt.Sprintf("Heads up! <20%% (%s) of your account quota remains",
				humanize.Bytes(uint64(free-totalSize))),
		)
	}
	return
}
