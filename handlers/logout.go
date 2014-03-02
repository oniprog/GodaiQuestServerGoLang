/*
   ログアウト処理を行う
*/
package handlers

import (
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"net/http"
)

// ログインの処理
func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	// ログインチェック
	client, err := sessions.GetClient(w, r)
	if err != nil {
		network.RedirectIndex(w, r, "", err.Error())
		return
	}

	// ログアウト処理
	client.Close()
	sessions.Logout(client)

	network.RedirectIndex(w, r, "", "ログアウトしました")
}
