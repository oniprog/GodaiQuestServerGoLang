package handlers

import (
	"errors"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"net/http"
	"strconv"
)

// 最後の記事の削除
func DeleteLastAritcleHandler(w http.ResponseWriter, r *http.Request) {

	// ログインチェック
	client, err := sessions.GetClient(w, r)
	if err != nil {
		network.RedirectIndex(w, r, "", err.Error())
		return
	}

	// POSTされたものかのチェック
	if r.Method != "POST" {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	// 情報番号
	infoId := 0
	queries := r.URL.Query()
	if len(queries["info_id"]) > 0 {
		infoId64, _ := strconv.ParseInt(queries["info_id"][0], 10, 0)
		infoId = int(infoId64)
	} else {
		err = errors.New("削除する対象の情報の指定がありません")
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	// 削除実行
	err = network.DeleteLastItemAritcle(client, infoId)
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	ReadInfoHandler(w, r)
}
