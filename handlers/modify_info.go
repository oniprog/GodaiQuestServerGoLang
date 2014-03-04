package handlers

import (
	"errors"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"github.com/oniprog/GodaiQuestServerGoLang/godaiquest"
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"net/http"
	"strconv"
)

// 記事の修正
func ModifyArticleHandler(w http.ResponseWriter, r *http.Request) {

	// ログインチェック
	client, err := sessions.GetClient(w, r)
	if err != nil {
		network.RedirectIndex(w, r, "", err.Error())
		return
	}

	// ページの表示用
	dataTemp := make(map[string]interface{})

	// 見るユーザ
	viewId := client.UserId
	queries := r.URL.Query()
	if len(queries["view_id"]) > 0 {
		dataTemp["view_id"] = queries["view_id"][0]
		viewId64, _ := strconv.ParseInt(queries["view_id"][0], 10, 0)
		viewId = int(viewId64)
	} else {
		err = errors.New("情報の指定がありません")
	}
	// 見る情報
	infoId := 0
	if len(queries["info_id"]) > 0 {
		dataTemp["info_id"] = queries["info_id"][0]
		infoId64, _ := strconv.ParseInt(queries["info_id"][0], 10, 0)
		infoId = int(infoId64)
	} else {
		err = errors.New("情報の指定がありません")
	}
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	// 情報を取り出す
	itemInfo, err := network.GetItemInfoByUserId(client, w, r, viewId)
	if err != nil {
		network.RedirectIndex(w, r, "", err.Error())
		return
	}
	var curItem *godaiquest.AItem
	for _, aitemdic := range itemInfo.GetAitemDic() {

		aitem := aitemdic.GetAitem()
		itemId := int(aitem.GetItemId())
		if itemId == infoId {

			curItem = aitem
			break
		}
	}
	if curItem == nil {
		network.RedirectIndex(w, r, "", "対象の情報がありません")
		return
	}
	dataTemp["aitem"] = curItem

	// POSTされたものかのチェック
	if r.Method == "POST" {

		// 書き込み内容
		newText := r.PostFormValue("inputtext")
		// 記事の書き込み
		network.ChangeAItem( client, int(curItem.GetItemId()), int(curItem.GetItemImageId()), newText )

		ReadInfoHandler(w, r)
	} else {

		// レンダリング
		template.Execute("modify_info", w, dataTemp)
	}
}
