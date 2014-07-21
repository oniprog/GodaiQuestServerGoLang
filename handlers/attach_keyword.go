package handlers

import (
	"fmt"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"net/http"
	"strconv"
)

func redirectAttachKeyword(w http.ResponseWriter, r *http.Request, message string, info_id int, view_id int) {

	redirectStr := fmt.Sprintf("/attach_keyword?message="+message+"&info_id=%d&view_id=%d", info_id, view_id)
	http.Redirect(w, r, redirectStr, http.StatusSeeOther)
}

// 記事の書き込み
func AttachKeywordHandler(w http.ResponseWriter, r *http.Request) {

	// ログインチェック
	client, err := sessions.GetClient(w, r)
	if err != nil {
		network.RedirectIndex(w, r, "", err.Error())
		return
	}

	// POSTされたものかのチェック
	if r.Method != "POST" {

		// ページの表示用
		dataTemp := make(map[string]interface{})

		queries := r.URL.Query()
		if len(queries["message"]) > 0 {
			dataTemp["message"] = queries["message"][0]
		}
		infoId := 0
		viewId := 0
		if len(queries["info_id"]) > 0 {
			dataTemp["info_id"] = queries["info_id"][0]
			infoId64, _ := strconv.ParseInt(queries["info_id"][0], 10, 0)
			infoId = int(infoId64)
		}
		if len(queries["view_id"]) > 0 {
			dataTemp["view_id"] = queries["view_id"][0]
			viewId64, _ := strconv.ParseInt(queries["view_id"][0], 10, 0)
			viewId = int(viewId64)
		}

		keywordUserInfo, err := network.ListKeyword(client, client.UserId)
		if err != nil {
			redirectAttachKeyword(w, r, err.Error(), infoId, viewId)
			return
		}

		// ソートしつつキーワードを並べる
		listKeyword := make(map[int]string, len(keywordUserInfo.GetKeywordSet()))
		for _, akeyword := range keywordUserInfo.GetKeywordSet() {
			priority := int(akeyword.GetKeywordPriority())
			for ; ; priority++ {
				_, ok := listKeyword[priority]
				if !ok {
					break
				}
			}
			listKeyword[priority] = akeyword.GetKeyword()
		}
		dataTemp["keyword_list"] = listKeyword

		// レンダリング
		template.Execute("attach_keyword", w, dataTemp)
	} else {

		dataTemp := make(map[string]interface{})

		keyword := r.PostFormValue("keyword")
		if len(keyword) == 0 {

			dataTemp["message"] = "空のキーワードではダメです"
			template.Execute("attach_keyword", w, dataTemp)
			return
		}

		view_id_str := r.PostFormValue("view_id")
		info_id_str := r.PostFormValue("info_id")

		if len(view_id_str) == 0 || len(info_id_str) == 0 {
			network.RedirectInfoTop(w, r, "", err.Error())
			return
		}

		viewId64, err := strconv.ParseInt(view_id_str, 10, 0)
		infoId64, err := strconv.ParseInt(info_id_str, 10, 0)
		viewId := int(viewId64)
		infoId := int(infoId64)
		if err != nil {
			redirectAttachKeyword(w, r, err.Error(), infoId, viewId)
			return
		}

		// キーワードのリストを得て該当するものを探す
		keywordUserInfo, err := network.ListKeyword(client, client.UserId)
		keywordId := -1
		itemPriority := 10000 - len(keywordUserInfo.GetKeywordSet())
		for _, akeyword := range keywordUserInfo.GetKeywordSet() {
			keyword2 := akeyword.GetKeyword()
			if keyword2 == keyword {
				keywordId = int(akeyword.GetKeywordId())
				break
			}
		}
		if keywordId < 0 {
			// キーワードが無いので新規作成する
			keywordId, err = network.RegisterKeyword(client, keyword)
			if err != nil {
				redirectAttachKeyword(w, r, "キーワードの新規作成に失敗しました", infoId, viewId)
				return
			}
		}

		// キーワードと記事をひもづける
		err = network.AttachKeyword(client, infoId, keywordId, itemPriority)
		if err != nil {
			dataTemp["message"] = err.Error()
			redirectAttachKeyword(w, r, err.Error(), infoId, viewId)
			return
		}

		// レンダリング
		//		http.Redirect(w, r, "/list_info?message="+message+"&view_id="+view_id_str+"&info_id="+info_id_str, http.StatusSeeOther)
		//		network.RedirectInfoTop(w,r, "", "キーワード登録できました"  )
		redirectAttachKeyword(w, r, "キーワードを登録できました. 窓を閉じてください", infoId, viewId)
	}
}
