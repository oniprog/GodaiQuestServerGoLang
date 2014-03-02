/*
　ログイン処理を行う
*/
package handlers

import (
	"net/http"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
)


// ログインの処理
func LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Redirect(w, r, "/index", http.StatusMovedPermanently)
		return
	}

    // 同じemailを持つアカウントの接続を切る
    email := r.PostFormValue("email")
    sessions.DeleteClientSameEmail( email )

    // 接続処理を行う
    client := network.TryLogon(w,r)

    // 登録処理
    if client != nil {
        email := r.PostFormValue("email")
        sessions.SetNewClient( w, r, client, email )
        http.Redirect(w, r, "/list_user", http.StatusSeeOther)
    }
}
