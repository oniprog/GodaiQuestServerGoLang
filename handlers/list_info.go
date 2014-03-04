package handlers

import (
	"github.com/oniprog/GodaiQuestServerGoLang/godaiquest"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"net/http"
	"strconv"
	"strings"
)

const pagesize = 10

// 情報の一覧（未読だけ)
func ListInfoHandler(w http.ResponseWriter, r *http.Request) {

	ListInfoHandlerCommon(false, w, r)
}

// 情報の一覧（全部)
func ListInfoAllHandler(w http.ResponseWriter, r *http.Request) {

	ListInfoHandlerCommon(true, w, r)
}

// 共通部分
func ListInfoHandlerCommon(all bool, w http.ResponseWriter, r *http.Request) {

	// ログインチェック
	client, err := sessions.GetClient(w, r)
	if err != nil {
		network.RedirectIndex(w, r, "", err.Error())
		return
	}

	// ページ用のデータ生成
	dataTemp := make(map[string]interface{})
	queries := r.URL.Query()
	if len(queries["message"]) > 0 {
		dataTemp["message"] = queries["message"][0]
	}
	// 見るユーザ
	viewId := client.UserId
	if len(queries["view_id"]) > 0 {
		dataTemp["view_id"] = queries["view_id"][0]
		viewId64, _ := strconv.ParseInt(queries["view_id"][0], 10, 0)
		viewId = int(viewId64)
	} else {
		dataTemp["view_id"] = string(client.UserId)
	}
	//
	index := -1
	if len(queries["index"]) > 0 {
		indexTmp, _ := strconv.ParseInt(queries["index"][0], 10, 0)
		index = int(indexTmp)
	}

	// すべてのユーザ情報の読み込み
	userInfo, err := network.GetAllUserInfo(client, w, r)
	if err != nil {
		network.RedirectLogonTop(w, r, "", err.Error())
		return
	}

	// ページごとの未読を読み込む
	mapUserUnread := make(map[int][]int)
	for _, auserdic := range userInfo.GetUesrDic() {

		auser := auserdic.GetAuser()
		dungeonId := int(auser.GetUserId())
		listUnread, _ := network.GetUnpickedupItemInfo(client, w, r, client.UserId, dungeonId)
		mapUserUnread[dungeonId] = listUnread

		if int(auser.GetUserId()) == viewId {
			// 名前を取り出す
			dataTemp["name"] = auser.GetUserName()
		}
	}

	// 未読情報を取り出す
	itemInfo, err := network.GetItemInfoByUserId(client, w, r, viewId)

	if err != nil {
		network.RedirectLogonTop(w, r, "", err.Error())
		return
	}

	// 変換する
	mapUnread := make(map[int]int)
	for _, id := range mapUserUnread[viewId] {
		mapUnread[int(id)] = 1
	}

	// 可視化用に調整する
	mapItem := make(map[int]*godaiquest.AItem)
	cntItem := 0
	for _, aitemdic := range itemInfo.GetAitemDic() {

		aitem := aitemdic.GetAitem()
		itemId := int(aitem.GetItemId())
		_, ok := mapUnread[itemId]
		if !ok && !all {
			continue
		}
		cntItem++
	}
	if index < 0 {
		if all {
			index = int(cntItem / pagesize)*pagesize
		} else {
			index = 0
		}
	}
	if index < 0 {
		index = 0
	}
	dataTemp["before"] = 0
	dataTemp["after"] = 0

	for i, aitemdic := range itemInfo.GetAitemDic() {

		aitem := aitemdic.GetAitem()
		itemId := int(aitem.GetItemId())
		_, ok := mapUnread[itemId]
		if !ok && !all {
			continue
		}
		if i < index {
			dataTemp["before"] = 1
		} else if i >= index && i <= index+pagesize {
			mapItem[i-index] = aitem
			strHeader := *aitem.HeaderString + "\n\n\n\n\n\n"
			newstr := strings.Join(strings.Split(strHeader, "\n")[0:5], "\n")
			aitem.HeaderString = &newstr
		} else {
			dataTemp["after"] = 1
		}
	}
	dataTemp["index"] = index

	dataTemp["itemlist"] = mapItem
	dataTemp["pagesize"] = pagesize

	// レンダリング
	if all {
		template.Execute("list_info_all", w, dataTemp)
	} else {
		template.Execute("list_info", w, dataTemp)
	}
}
