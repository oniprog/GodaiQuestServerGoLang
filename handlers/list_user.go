package handlers

import (
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"net/http"
)

// 処理
func ListUserHandler(w http.ResponseWriter, r *http.Request) {

	// ログインチェック
	client, err := sessions.GetClient(w, r)
	if err != nil {
		network.RedirectIndex(w, r, "", err.Error())
		return
	}

	// すべてのユーザ情報の読み込み
	userInfo, err := network.GetAllUserInfo(client, w, r)
	if err != nil {
		network.RedirectIndex(w, r, "", err.Error())
		return
	}

	// ページごとの未読を読み込む
	mapUserUnread := make(map[int][]int)
	for _, auserdic := range userInfo.GetUesrDic() {

		auser := auserdic.GetAuser()
		dungeonId := int(auser.GetUserId())
		listUnread, _ := network.GetUnpickedupItemInfo(client, w, r, client.UserId, dungeonId)
		mapUserUnread[dungeonId] = listUnread
	}
	// ページ用のデータ生成
	dataTemp := make(map[string]interface{})

	// ユーザ名と画像などの生成
	mapUsers := make(map[int]interface{})
	for _, auserdic := range userInfo.GetUesrDic() {

		auser := auserdic.GetAuser()
		mapUser := make(map[string]interface{})
		mapUser["UserId"] = auser.GetUserId()
		mapUser["EMail"] = auser.GetMailAddress()
		mapUser["UserName"] = auser.GetUserName()
		mapUser["UriImage"] = string(network.ConvURIImage(auser.GetUserImage()))
		mapUser["UnreadCount"] = len(mapUserUnread[int(auser.GetUserId())])
		mapUsers[int(auser.GetUserId())] = mapUser
	}
	dataTemp["UserInfo"] = mapUsers

	// メッセージを取り出しておく
	queries := r.URL.Query()
	if len(queries["message"]) > 0 {
		dataTemp["message"] = queries["message"][0]
	}

	// レンダリング
	template.Execute("list_user", w, dataTemp)
}
