package web

import (
	"encoding/base64"
	"net/http"
	"time"
)

func setMessage(w http.ResponseWriter, name, value string) {
	mesageCookie := http.Cookie{
		Name:  name,
		Value: base64.StdEncoding.EncodeToString([]byte(value)),
		Path:  "/",
	}

	http.SetCookie(w, &mesageCookie)
}

func getMessage(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	messageCookie, err := r.Cookie(name)
	if err != nil {
		switch err {
		case http.ErrNoCookie:
			return "", nil
		default:
			return "", err
		}
	}

	clearCookie := http.Cookie{
		Name:    name,
		MaxAge:  -1,
		Expires: time.Unix(1, 0),
		Path:    "/",
	}

	http.SetCookie(w, &clearCookie)

	decodedValue, _ := base64.URLEncoding.DecodeString(messageCookie.Value)

	return string(decodedValue), nil
}
