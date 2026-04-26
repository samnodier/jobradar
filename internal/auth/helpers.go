package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
)

func redirectWithError(w http.ResponseWriter, r *http.Request, path, code string) {
	http.Redirect(w, r, path+"?error="+url.QueryEscape(code), http.StatusFound)
}

func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
