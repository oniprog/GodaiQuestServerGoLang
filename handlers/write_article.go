package handlers

import (
	"errors"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"net/http"
	"strconv"
)

// 情報内の記事の投稿
func WriteArticleHandler(w http.ResponseWriter, r *http.Request) {

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
		err = errors.New("書き込む情報の指定がありません")
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	// 書き込む内容
	contents := r.PostFormValue("inputtext")

	err = network.SetItemArticle( client, infoId, 0, client.UserId, contents )
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}
	
	ReadInfoHandler(w, r)
}
