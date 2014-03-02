package sessions

import (
    "github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("i"))

func Prepare(secret string) error {
     store = sessions.NewCookieStore([]byte(secret))
     return nil
}
