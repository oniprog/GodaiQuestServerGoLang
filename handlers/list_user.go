package handlers

import (
    "net/http"
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
)

// 処理
func ListUserHandler(w http.ResponseWriter, r *http.Request) {

    _, err := sessions.GetClient( w, r )
    if err != nil {
        RedirectIndex( w, r, "", err.Error() )
        return
    }
	data2 := make(map[string]string)
    dataall := []interface{}{ data2 }
	template.Execute("list_user", w, dataall)
}