package main

import (
	"fmt"
	"gitern/account"
	"gitern/art"
	"os"
	"strings"
)

func main() {
	if len(os.Args) > 2 {
		art.Fool.Fatal("Usage: gitern-account [<account name>]")
	}

	fp := os.Getenv("FP")
	var accname string

	if len(os.Args) == 1 {
		accNames, err := account.AccountNames()
		if err != nil {
			art.Tower.Fatal(err.Error())
		}
		if len(accNames) > 1 {
			art.Fool.Fatal("Must specify account name.", strings.Join(accNames, " or "))
		}

		accname = accNames[0]
	} else {
		accname = os.Args[1]
	}

	id, err := account.CreateSession(accname, fp)
	if err != nil {
		art.Tower.Fatal(err.Error())
	}

	art.Sun.Print(
		fmt.Sprintf("View account %s at gitern.com/%s in a web browser.", accname, id),
		"This link expires in 5 minutes and is only good for one use.")
}
