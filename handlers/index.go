/*
   ログイン画面を扱うためのハンドラー
*/
package handlers

import (
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"net/http"
)

// 処理
func IndexHandler(w http.ResponseWriter, r *http.Request) {

	data := make(map[string]string)

	queries := r.URL.Query()
	if len(queries["message"]) > 0 {
		data["message"] = queries["message"][0]
	}
	if len(queries["email"]) > 0 {
		data["email"] = queries["email"][0]
	}

	template.Execute("index", w, data)
}
