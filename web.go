package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"html/template"
	"net/http"

	"gitern/db"
	"gitern/pubkey"

	"github.com/dpapathanasiou/go-recaptcha"
	"github.com/joho/godotenv"
)

const GITERN_ACCOUNTS = "/gitern/accounts/"

func isDev() bool {
	return os.Getenv("ENV") == "DEV"
}

var funcMap = template.FuncMap{
	"isDev": isDev,
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env found ... using default env")
	}

	schema, err := ioutil.ReadFile("schema.sql")
	if err != nil {
		log.Fatalln(err)
	}

	_, err = db.Conn.Exec(string(schema))
	if err != nil {
		log.Fatalln(err)
	}

	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = "5000"
	}

	http.Handle("/public/", http.StripPrefix("/public/",
		http.FileServer(http.Dir("./public"))))
	http.HandleFunc("/", index)
	http.HandleFunc("/join", join)
	http.HandleFunc("/account", account)

	stripeHTTP()

	log.Printf("starting gitern server on port %s\n\n", port)
	log.Fatalln(http.ListenAndServe(":"+port, nil))
}

func session(w http.ResponseWriter, r *http.Request) {
	// check if the ssh based session exists
	row := db.Conn.QueryRow(`SELECT name, fingerprint, 
							 NOW() - created_at > interval '1 hours' AS expired 
							 FROM sessions 
							 WHERE id = $1`, strings.TrimPrefix(r.URL.Path, "/"))
	var name, fp string
	var expired bool
	if err := row.Scan(&name, &fp, &expired); err != nil {
		errHandlerMsg(w, http.StatusUnauthorized, "invalid session")
		return
	}
	db.Conn.Exec("SELECT delete_sessions($1, $2)", name, fp)

	if expired {
		errHandlerMsg(w, http.StatusUnauthorized, "expired session")
		return
	}

	err := createSession(w, name, fp)
	if err != nil {
		errHandler(w, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		session(w, r)
		return
	}

	tmpl := template.Must(template.New("index.tmpl").Funcs(funcMap).
		ParseFiles("views/index.tmpl"))
	tmpl.Execute(w, nil)
}

type tArgs map[string]interface{}

func pubkeySplit(pubkey string) (string, string, string) {
	parts := strings.Split(pubkey, " ")
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

func account(w http.ResponseWriter, r *http.Request) {
	accname, fp, err := getSession(r)
	if err != nil {
		errHandler(w, http.StatusUnauthorized)
		return
	}

	stripeID, stripeStatus, err := stripeIDAndStatus(accname)
	if err != nil {
		errHandlerMsg(w, http.StatusInternalServerError, "retrieving stripe status")
		return
	}

	keyCount, _ := pubkey.Count(accname)

	tmpl := template.Must(template.ParseFiles("views/account.tmpl"))
	tmpl.Execute(w, tArgs{
		"accname":                accname,
		"fp":                     fp,
		"pubkeys":                keyCount,
		"StripeID":               stripeID,
		"StripeStatus":           stripeStatus,
		"STRIPE_PUBLISHABLE_KEY": os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		"STRIPE_PRICE_ID":        os.Getenv("STRIPE_PRICE_ID"),
	})
}

func join(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm() err: %v", err)
		errHandler(w, http.StatusInternalServerError)
		return
	}

	chaKey := os.Getenv("RECAPTCHA_SECKEY")
	if chaKey == "" {
		log.Println("no recaptcha secret key")
		errHandler(w, http.StatusInternalServerError)
		return
	}
	recaptcha.Init(chaKey)
	chaResp := r.FormValue("g-recaptcha-response")
	result, err := recaptcha.Confirm(r.RemoteAddr, chaResp)
	if err != nil {
		log.Println("recaptcha confirmation error", err)
		errHandler(w, http.StatusInternalServerError)
		return
	}
	if !result && !isDev() {
		errHandler(w, http.StatusBadRequest)
		return
	}

	accname := r.FormValue("accname")
	err = db.DoTxn(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO accounts (name) VALUES ($1)", accname)
		if err != nil {
			return err
		}

		err = pubkey.AddLineTx(r.FormValue("pubkey"), accname, tx)
		if err != nil {
			return err
		}

		err = os.Mkdir(path.Join(GITERN_ACCOUNTS, accname), 0755)
		if err != nil && !isDev() {
			return err
		}

		return nil
	})
	if err != nil {
		log.Printf("sign up err: %v", err)
		errHandler(w, http.StatusBadRequest)
		return
	}

	row := db.Conn.QueryRow(`SELECT pubkey_fingerprint
							 FROM accounts_pubkeys 
							 WHERE account_name = $1`, accname)
	var fp string
	if err := row.Scan(&fp); err != nil {
		log.Printf("row.Scan() err: %v", err)
		errHandler(w, http.StatusInternalServerError)
		return
	}

	err = createSession(w, accname, fp)
	if err != nil {
		errHandler(w, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

func errHandler(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	fmt.Fprintf(w, "Error  %d\n", status)
}

func errHandlerMsg(w http.ResponseWriter, status int, msg string) {
	errHandler(w, status)
	fmt.Fprintf(w, msg)
}
