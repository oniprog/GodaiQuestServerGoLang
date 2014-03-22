package handlers

import (
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"net/http"
	"strconv"
	"fmt"
)


func redirectManageKeyword(w http.ResponseWriter, r * http.Request, message string, view_id int, keyword string ) {

	redirectStr := fmt.Sprintf("/list_info_by_keyword?message="+message+"&keyword="+keyword+"&view_id=%d", view_id )
	http.Redirect(w, r, redirectStr, http.StatusSeeOther)
}

func ManageKeywordHandler(w http.ResponseWriter, r *http.Request) {

	// ログインチェック
	client, err := sessions.GetClient(w, r)
	if err != nil {
		network.RedirectIndex(w, r, "", err.Error())
		return
	}

	// 
	dataTemp := make(map[string]interface{})
	queries := r.URL.Query()
	if len(queries["message"]) > 0 {
		dataTemp["message"] = queries["message"][0]
	}
	
	// 見るユーザ
	info_id_str := r.PostFormValue("info_id")
	view_id_str := r.PostFormValue("view_id")
	if len(view_id_str) == 0 {
		redirectManageKeyword(w,r,"不正な操作です", 0, "")
		return
	}
	infoId64,_ := strconv.ParseInt( info_id_str, 10, 0 )
	viewId64,_ := strconv.ParseInt( view_id_str, 10, 0 )
	infoId := int(infoId64)
	viewId := int(viewId64)

	// キーワードを取り出す
	keyword := r.PostFormValue("keyword")
	dataTemp["keyword"] = keyword
	if viewId != client.UserId {
		redirectManageKeyword(w,r,"自分のキーワードしか操作できません", viewId, keyword )
		return
	}

	if r.Method != "POST" {
		redirectManageKeyword(w,r,"不正な操作です", viewId, keyword)
		return
	}


	// キーワードIdを得る
	keywordUserInfo, err := network.ListKeyword(client, viewId)
	if err != nil {
		network.RedirectLogonTop(w, r, "", err.Error())
		return
	}
	keywordId := -1
	for _, akeyword := range keywordUserInfo.GetKeywordSet() {
		if akeyword.GetKeyword() == keyword {
			keywordId = int(akeyword.GetKeywordId())
			break
		}
	}
	if keywordId < 0 {
		redirectManageKeyword(w,r, "存在しないキーワードです", viewId, keyword)
		return
	}

	// 処理の内容を分ける
	remove_info := r.PostFormValue("remove_info")
	if len(remove_info) != 0 {

		if len(view_id_str) == 0 {
			redirectManageKeyword(w,r,"不正な操作です", 0, "")
			return
		}
		// 記事を外す
		err := network.DetachKeyword( client, infoId, keywordId )
		if err != nil {

			redirectManageKeyword(w,r, err.Error(), viewId, keyword)
			return
		}

		redirectManageKeyword(w,r, "記事を外しました", viewId, keyword)
		return
	}
	delete_keyword := r.PostFormValue("delete_keyword")
	if len(delete_keyword) > 0 {

		// キーワードを消す
		err := network.DeleteKeyword( client, keywordId )
		if err != nil {
			redirectManageKeyword(w,r, err.Error(), viewId, keyword)
			return
		}

		network.RedirectLogonTop(w,r, "", "キーワードを消しました")
		return
	}
	redirectManageKeyword(w,r, "内部エラー", viewId, keyword)
}

