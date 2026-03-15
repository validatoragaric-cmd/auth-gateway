package authgateway

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
)

const (
	sessionCookieName = "auth-gateway-session"
	sessionTimeout    = 30 * time.Minute
)

var (
	cookieHashKey  = []byte("your_secret_hash_key")
	cookieBlockKey = []byte("your_secret_block_key")
	cookieEncoder  = securecookie.New(cookieHashKey, cookieBlockKey)
)

func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func hashPassword(password string) (string, error) {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil)), nil
}

func verifyPassword(storedPassword, providedPassword string) bool {
	h := sha256.New()
	h.Write([]byte(providedPassword))
	return subtle.ConstantTimeCompare([]byte(storedPassword), h.Sum(nil)) == 1
}

func setSessionCookie(w http.ResponseWriter, sessionID string) error {
	encodedSession, err := cookieEncoder.Encode(sessionCookieName, sessionID)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:     encodedSession,
		MaxAge:    int(sessionTimeout.Seconds()),
		Secure:    true,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	return nil
}

func getSessionCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}
	var sessionID string
	err = cookieEncoder.Decode(sessionCookieName, cookie.Value, &sessionID)
	if err != nil {
		return "", err
	}
	return sessionID, nil
}

func isAuthorized(r *http.Request) bool {
	sessionID, err := getSessionCookie(r)
	if err != nil {
		log.Println(err)
		return false
	}
	if sessionID == "" {
		return false
	}
	return true
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func getBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header found")
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization header")
	}
	return parts[1], nil
}