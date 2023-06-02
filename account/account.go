package account

import (
	"fmt"
	"gitern/db"
	"os"
	"path/filepath"
	"strings"
)

const ACCOUNTS_PATH = "/accounts"

type Account struct {
	Name   string
	Status string
	Quota  int64
}

// path before being chrooted
func (account Account) Path() (string, error) {
	return AccountPath(account.Name)
}

// path after being chroot'd
func (account Account) ReposPath() string {
	return filepath.Join("/", account.Name)
}

const MEGABYTE = 1024 * 1024

func Accounts() ([]Account, error) {
	authAccounts := os.Getenv("ACCOUNT")
	if authAccounts == "" {
		return nil, fmt.Errorf("You don't administer any accounts")
	}

	var accounts []Account
	authAccountsSplit := strings.Split(authAccounts, ",")
	for _, authAccountStr := range authAccountsSplit {
		var account Account
		_, err := fmt.Sscan(authAccountStr, &account.Name, &account.Status)
		if err != nil {
			return nil, err
		}
		if account.Status == "paid" {
			account.Quota = 0
		} else {
			// TODO: take from env
			account.Quota = 25 * MEGABYTE
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func AccountNames() ([]string, error) {
	accounts, err := Accounts()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, account := range accounts {
		names = append(names, account.Name)
	}

	return names, nil
}

func CheckAccess(accountName string) error {
	_, err := GetAuthAccount(accountName)
	return err
}

func GetAuthAccount(accountName string) (Account, error) {
	authAccounts, err := Accounts()
	if err != nil {
		return Account{}, err
	}

	for _, a := range authAccounts {
		if a.Name == accountName {
			return a, nil
		}
	}

	return Account{}, fmt.Errorf("Unauthorized for account %s", accountName)
}

func AccountPath(accountName string) (string, error) {
	path := filepath.Join(ACCOUNTS_PATH, accountName)
	_, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	return path, nil
}

func CreateSession(accountName string, fp string) (string, error) {
	err := CheckAccess(accountName)
	if err != nil {
		return "", err
	}

	row := db.Conn.QueryRow(`SELECT create_session($1, $2)`, accountName, fp)

	var id string
	if err := row.Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}
