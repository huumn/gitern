package pubkey

import (
	"gitern/db"
	"strings"
)

func splitLine(line string) (string, string, string) {
	parts := strings.Split(line, " ")
	switch len(parts) {
	case 0:
		return "", "", ""
	case 1:
		return parts[0], "", ""
	case 2:
		return parts[0], parts[1], ""
	default:
		return parts[0], parts[1], parts[2]
	}
}

func AddLineTx(line, account string, tx db.DBTX) error {
	keytype, key, comment := splitLine(line)

	// if they pubkey already exists, continue ie do nothing
	_, err := tx.Exec(`INSERT INTO pubkeys (fingerprint, keytype, pubkey, comment) 
					   VALUES (pubkey2fingerprint($2), $1, $2, $3) 
					   ON CONFLICT ON CONSTRAINT pubkeys_pkey DO NOTHING`,
		keytype, key, comment)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO accounts_pubkeys (account_name, pubkey_fingerprint) 
					  VALUES ($1, pubkey2fingerprint($2))`, account, key)
	if err != nil {
		return err
	}

	return nil
}

type Key struct {
	KeyType string
	Pubkey  string
	Comment string
	Fp      string
}

func ListTx(account string, tx db.DBTX) ([]Key, error) {
	var keys []Key

	rows, err := tx.Query(`SELECT fingerprint, keytype, pubkey, comment
								FROM pubkeys
								JOIN accounts_pubkeys 
								ON fingerprint = accounts_pubkeys.pubkey_fingerprint
								WHERE accounts_pubkeys.account_name = $1`, account)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key Key
		if err := rows.Scan(&key.Fp, &key.KeyType, &key.Pubkey, &key.Comment); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}

func List(account string) ([]Key, error) {
	return ListTx(account, db.Conn)
}

func CountTx(account string, tx db.DBTX) (int, error) {
	keys, err := ListTx(account, tx)
	if err != nil {
		return 0, err
	}

	return len(keys), err
}

func Count(account string) (int, error) {
	return CountTx(account, db.Conn)
}

func RemoveFingerprintTx(account, fp string, tx db.DBTX) error {
	_, err := tx.Exec(`DELETE FROM accounts_pubkeys
					   WHERE account_name = $1
					   AND pubkey_fingerprint = $2`, account, fp)
	if err != nil {
		return err
	}

	return nil
}

func RemoveFingerprint(account, fp string) error {
	return RemoveFingerprintTx(account, fp, db.Conn)
}
