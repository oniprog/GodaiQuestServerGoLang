package handlers

import (
    "net/http"
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
)

// 処理
func ListUserHandler(w http.ResponseWriter, r *http.Request) {

     // ログインチェック
    _, err := sessions.GetClient( w, r )
    if err != nil {
        network.RedirectIndex( w, r, "", err.Error() )
        return
    }

	dataMap := make(map[string]string)
	template.Execute("list_user", w, dataMap)
}