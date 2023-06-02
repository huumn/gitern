package main

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	secKey        = "cookie.key" // openssl genrsa -out cookie.key
	pubKey        = "cookie.pub" // openssl rsa -in cookie.key -pubout > cookie.pub
	sessionLength = 1 * time.Hour
	sessionName   = "session"
)

func decryptToken(encToken string) (string, string, error) {
	token, err := jwt.ParseWithClaims(encToken, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})
	if err != nil {
		return "", "", err
	}

	claims := token.Claims.(*AdminClaims)

	return claims.Account, claims.Fingerprint, nil
}

func getSession(r *http.Request) (string, string, error) {
	cookie, err := r.Cookie(sessionName)
	if err != nil || cookie == nil {
		return "", "", fmt.Errorf("No session set available")
	}

	return decryptToken(cookie.Value)
}

func setEncryptedSession(w http.ResponseWriter, encryptedToken string) {
	setCookie(w, sessionName, encryptedToken, int(sessionLength.Seconds()))
}

func createSession(w http.ResponseWriter, account, fingerprint string) error {
	return createSessionLength(w, account, fingerprint, sessionLength)
}

func createSessionLength(w http.ResponseWriter, account, fingerprint string,
	length time.Duration) error {
	tokenString, err := createToken(account, fingerprint, length)
	if err != nil {
		return err
	}

	setCookie(w, sessionName, tokenString, int(length.Seconds()))
	return err
}

func setCookie(w http.ResponseWriter, name string, value string, maxAge int) {
	c := &http.Cookie{Name: name, MaxAge: maxAge, Value: value, HttpOnly: true, Path: "/"}
	http.SetCookie(w, c)
}

// AdminClaims JWT data
type AdminClaims struct {
	jwt.StandardClaims
	Account     string `json:"account"`
	Fingerprint string `json:"fingerprint"`
}

func createToken(account, fingerprint string, dur time.Duration) (string, error) {
	t := jwt.New(jwt.GetSigningMethod("RS256"))

	t.Claims = AdminClaims{
		jwt.StandardClaims{
			// see http://tools.ietf.org/html/draft-ietf-oauth-json-web-token-20#section-4.1.4
			// 1 month from now
			ExpiresAt: time.Now().Add(dur).Unix(),
		},
		account,
		fingerprint,
	}

	return t.SignedString(signKey)
}

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

func init() {
	secKey, exists := os.LookupEnv("SECKEY_PATH")
	if !exists {
		/* we are in local dev mode */
		secKey = "cookie.key" // openssl genrsa -out jwt.key
	}
	signBytes, err := ioutil.ReadFile(secKey)
	if err != nil {
		log.Fatal(err)
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatal(err)
	}
	pubKey, exists := os.LookupEnv("PUBKEY_PATH")
	if !exists {
		/* we are in local dev mode */
		pubKey = "cookie.pub" // openssl genrsa -out jwt.key
	}
	verifyBytes, err := ioutil.ReadFile(pubKey)
	if err != nil {
		log.Fatal(err)
	}
	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatal(err)
	}
}
