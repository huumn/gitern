package main

import (
	"bufio"
	"fmt"
	"gitern/account"
	"gitern/art"
	"gitern/pubkey"
	"io"
	"os"
)

func printKeyList(w io.Writer, accName string, full bool) {
	fp := os.Getenv("FP")

	err := account.CheckAccess(accName)
	if err != nil {
		art.Fool.Fatal(err.Error())
	}

	keys, err := pubkey.List(accName)
	if err != nil {
		art.Tower.Fatal("listing pubkeys", err.Error())
	}

	for _, k := range keys {
		var you string
		if k.Fp == fp {
			you = " (you)"
		}
		if full {
			fmt.Fprintf(w, "%s %s %s\n", k.KeyType, k.Pubkey, k.Comment)
		} else {
			fmt.Fprintf(w, "SHA256:%s %s%s\n", k.Fp, k.Comment, you)
		}
	}
}

func main() {
	full := false
	for i, v := range os.Args {
		if v == "--full" {
			full = true
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
		}
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	if len(os.Args) == 1 {
		accNames, err := account.AccountNames()
		if err != nil {
			art.Tower.Fatal("get account", err.Error())
		}
		for i, accName := range accNames {
			fmt.Fprintf(w, "%s:\n", accName)
			printKeyList(w, accName, full)
			if i != len(accNames)-1 {
				fmt.Fprintln(w)
			}
		}
	} else if len(os.Args) == 2 {
		printKeyList(w, os.Args[1], full)
	} else {
		art.Fool.Fatal("Usage: gitern-pubkey-list [--full] [<account>]")
	}
}
