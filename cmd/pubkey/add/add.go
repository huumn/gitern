package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"gitern/account"
	"gitern/art"
	"gitern/db"
	"gitern/pubkey"
	"gitern/stripehelper"
	"os"

	"github.com/omeid/pgerror"
)

func main() {
	if len(os.Args) != 2 {
		art.Fool.Fatal("Usage: gitern-pubkey-add <account>")
	}

	// we recheck their authorization for account just in case
	// something slips by Intake's auth checks
	accName := os.Args[1]
	err := account.CheckAccess(accName)
	if err != nil {
		art.Fool.Fatal(err.Error())
	}

	// TODO: is this a DOS attack vector if we sit waiting
	// to read from stdin?
	err = db.DoTxn(func(tx *sql.Tx) error {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			err := pubkey.AddLineTx(scanner.Text(), accName, tx)
			if err != nil {
				switch {
				case pgerror.UniqueViolation(err) != nil:
					art.Fool.Fatal(accName, "pubkey already exists")
				case pgerror.InvalidTextRepresentation(err) != nil:
					art.Fool.Fatal(accName, "invalid pubkey type", "should begin with ssh-rsa, ssh-ed25519, etc.")
				case pgerror.CheckViolation(err) != nil:
					art.Fool.Fatal(accName, "invalid pubkey representation")
				default:
					art.Tower.Fatal("add line", err.Error())
				}
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}

		return stripehelper.ReportUsageTx(accName, tx)
	})
	if err != nil {
		art.Tower.Fatal("pubkey add", err.Error())
	}

	art.Scroll.Print(fmt.Sprintf("pubkeys granted access to %s", accName))
}
