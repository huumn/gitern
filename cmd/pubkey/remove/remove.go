package main

import (
	"database/sql"
	"fmt"
	"gitern/account"
	"gitern/art"
	"gitern/db"
	_ "gitern/logmill"
	"gitern/pubkey"
	"gitern/stripehelper"
	"os"
	"strings"
)

const FP_PREFIX = "SHA256:"

func main() {
	if len(os.Args) != 3 {
		art.Fool.Fatal("Usage: gitern-pubkey-remove <fingerprint> <account>")
	}

	accountName := os.Args[2]
	err := account.CheckAccess(accountName)
	if err != nil {
		art.Fool.Fatal(err.Error())
	}

	fp := os.Args[1]

	if !strings.HasPrefix(fp, FP_PREFIX) {
		art.Fool.Fatal(fp, "Fingerprint expected to begin with "+FP_PREFIX)
	}

	fp = strings.TrimPrefix(fp, FP_PREFIX)

	err = db.DoTxn(func(tx *sql.Tx) error {
		err = pubkey.RemoveFingerprintTx(accountName, fp, tx)
		if err != nil {
			return err
		}

		return stripehelper.ReportUsageTx(accountName, tx)
	})
	if err != nil {
		art.Tower.Fatal(fp, "removing pubkey", err.Error())
	}

	art.Noose.Print(FP_PREFIX+fp, fmt.Sprintf("access revoked from %s", accountName))
}
