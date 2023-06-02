package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var Conn *sql.DB

// These rds can optionally be passed in at the linker stage
// We do this because we want binaries that don't have access
// to eb env to have these values when run, e.g. authorized-keys
// https://stackoverflow.com/questions/28459102/golang-compile-environment-variable-into-binary
// https://golang.org/cmd/link/
var (
	RDS_DB_NAME  string
	RDS_USERNAME string
	RDS_PASSWORD string
	RDS_HOSTNAME string
	RDS_PORT     string
)

type DBTX interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Prepare(string) (*sql.Stmt, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
}

// If env vars are available override default or linker values
func override() {
	str, exists := os.LookupEnv("RDS_DB_NAME")
	if exists {
		RDS_DB_NAME = str
	}
	str, exists = os.LookupEnv("RDS_USERNAME")
	if exists {
		RDS_USERNAME = str
	}
	str, exists = os.LookupEnv("RDS_PASSWORD")
	if exists {
		RDS_PASSWORD = str
	}
	str, exists = os.LookupEnv("RDS_HOSTNAME")
	if exists {
		RDS_HOSTNAME = str
	}
	str, exists = os.LookupEnv("RDS_PORT")
	if exists {
		RDS_PORT = str
	}
}

func dsn() string {
	var dsn string

	override()

	if RDS_DB_NAME == "" {
		/* we are in local dev mode */
		RDS_DB_NAME = "gitern sslmode=disable"
	}

	envs := map[string]string{
		"database": RDS_DB_NAME,
		"user":     RDS_USERNAME,
		"password": RDS_PASSWORD,
		"host":     RDS_HOSTNAME,
		"port":     RDS_PORT,
	}
	for k, v := range envs {
		if v != "" {
			dsn = dsn + " " + k + "=" + v
		}
	}

	return dsn
}

type Txn func(tx *sql.Tx) error

func DoTxn(fn Txn) error {
	tx, err := Conn.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		xerr := tx.Rollback()
		if xerr != nil {
			return xerr
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func init() {
	var err error
	Conn, err = sql.Open("postgres", dsn())
	if err != nil {
		log.Fatalln(err)
	}
}
