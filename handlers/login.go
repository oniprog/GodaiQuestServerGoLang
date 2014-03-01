package handlers

import (
	"net/http"
)

// ログインの処理
func LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Redirect(w, r, "/index", http.StatusMovedPermanently)
		return
	}

	http.Redirect(w, r, "/index?message=ログインしました", http.StatusSeeOther)
}
