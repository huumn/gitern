package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gitern/db"
)

const FP_PREFIX = "SHA256:"

func main() {
	if len(os.Args) != 4 {
		log.Fatalln("Usage: gitern-authorized-keys <fp> <key-type> <key>")
	}

	fingerprint := os.Args[1]
	if !strings.HasPrefix(fingerprint, FP_PREFIX) {
		log.Fatalln("Fingerprint must begin with " + FP_PREFIX)
	}

	fingerprint = strings.TrimPrefix(fingerprint, FP_PREFIX)

	// e.g. array_to_string: keyan paid,k00b unpaid,kk failed
	row := db.Conn.QueryRow(`SELECT array_to_string(array_agg(concat(name, ' ', COALESCE(text(stripe_status),'unpaid'))), ','),
							 fingerprint, keytype, pubkey
							 FROM pubkeys
							 JOIN accounts_pubkeys
							 ON fingerprint = pubkey_fingerprint
							 JOIN accounts
							 ON name = account_name
							 WHERE fingerprint = $1
							 GROUP BY fingerprint`, fingerprint)

	var account, fp, keytype, pubkey string
	if err := row.Scan(&account, &fp, &keytype, &pubkey); err != nil {
		fmt.Printf("environment=\"FP=%s\",restrict %s %s\n", fingerprint, os.Args[2], os.Args[3])
		return
	}

	fmt.Printf("environment=\"ACCOUNT=%s\",environment=\"FP=%s\",restrict %s %s\n", account, fp, keytype, pubkey)
}
