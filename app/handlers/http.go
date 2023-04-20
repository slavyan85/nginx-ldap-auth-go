package handlers

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/vkryuchenko/nginx-ldap-auth-go/app/clients"
	"go.uber.org/zap"
)

type HttpHandler struct {
	CookieName string
	LdapClient *clients.LdapClient
	Logger     *zap.Logger
}

func (*HttpHandler) DefaultRoute(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func (h *HttpHandler) Error(message string, w http.ResponseWriter, r *http.Request) {
	h.Logger.Warn(message, zap.String("ip", getClientIP(r)))
	w.Header().Set("WWW-Authenticate", "Basic realm=None")
	w.Header().Set("Cache-Control", "no-cache")
	http.Error(w, "", http.StatusUnauthorized)
}

func (h *HttpHandler) AuthRoute(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	authCokie, err := r.Cookie(h.CookieName)
	switch err {
	case http.ErrNoCookie:
		authCokie = &http.Cookie{
			Name:    h.CookieName,
			Value:   "",
			Expires: time.Now().AddDate(0, 0, 30),
		}
	case nil:
		break
	default:
		h.Error(err.Error(), w, r)
		return
	}
	if authCokie.Value != "" {
		authHeader = "Basic " + authCokie.Value
	}
	if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "basic ") {
		h.Error("authHeader empty or not contain requirement prefix", w, r)
		return
	}
	user, password, err := decodeAuth(authHeader)
	if err != nil {
		h.Error(err.Error(), w, r)
		return
	}
	_, _, err = h.LdapClient.Authenticate(user, password)
	if err != nil {
		h.Error(err.Error(), w, r)
		return
	}
	authCokie.Value = base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{user, password}, ":")))
	http.SetCookie(w, authCokie)
	w.WriteHeader(http.StatusOK)
}

func decodeAuth(auth string) (string, string, error) {
	decoded, err := base64.StdEncoding.DecodeString(auth[6:])
	if err != nil {
		return "", "", err
	}
	splited := strings.Split(string(decoded), ":")
	switch {
	case len(splited) == 2:
		return splited[0], splited[1], nil
	case len(splited) > 2:
		return splited[0], strings.Join(splited[1:], ":"), nil
	default:
		return "", "", errors.New(auth + " is not auth data")
	}
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return strings.Split(ip, ":")[0]
}
