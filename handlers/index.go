package handlers

import (
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"net/http"
)

// '/' の処理
func IndexHandler(w http.ResponseWriter, r *http.Request) {

	data := make(map[string]string)

	queries := r.URL.Query()
	if len(queries["message"]) > 0 {
		data["message"] = queries["message"][0]
	}
	email := r.Form["email"]
	password := r.Form["password"]

	if len(email) > 0 && len(password) > 0 {
		data["email"] = email[0]
		data["password"] = password[0]
	}
	template.Execute("index", w, data)
}
