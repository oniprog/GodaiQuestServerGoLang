package handlers

import (
	"fmt"
	//"github.com/oniprog/GodaiQuestServerGoLang/godaiquest"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"net/http"
	"strconv"
	"strings"
    "time"
)

// C#側でticksで得た値を補正するための定数
const TIME_CONST = -6795364578

func DateTimeString(cur int64) string {

    return time.Unix( TIME_CONST, cur*100).String()[0:19]
}

type aitem3 struct {
    ItemId int32
    HeaderString *string
    Created string
    LastModified string
}

func redirectListInfoByKeyword(w http.ResponseWriter, r *http.Request, message string, info_id int, view_id int, keyword string) {

	redirectStr := fmt.Sprintf("/list_info_by_keyword?message="+message+"&info_id=%d&keyword="+keyword+"&view_id=%d", info_id, view_id)
	http.Redirect(w, r, redirectStr, http.StatusSeeOther)
}

func ListInfoByKeywordHandler(w http.ResponseWriter, r *http.Request) {

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

	// キーワードを取り出す
	keyword := ""
	if len(queries["keyword"]) > 0 {
		keyword = queries["keyword"][0]
	} else {
		network.RedirectIndex(w, r, "", err.Error())
	}
	dataTemp["keyword"] = keyword

	// 見るユーザ
	viewId := client.UserId
	if len(queries["view_id"]) > 0 {
		dataTemp["view_id"] = queries["view_id"][0]
		viewId64, _ := strconv.ParseInt(queries["view_id"][0], 10, 0)
		viewId = int(viewId64)
	} else {
		dataTemp["view_id"] = string(client.UserId)
	}

	// 操作可能かどうか
	dataTemp["can_manip"] = viewId == client.UserId

	// すべてのユーザ情報の読み込み
	userInfo, err := network.GetAllUserInfo(client, w, r)
	if err != nil {
		network.RedirectLogonTop(w, r, "", err.Error())
		return
	}
	// ページごとの未読を読み込む
	for _, auserdic := range userInfo.GetUesrDic() {

		auser := auserdic.GetAuser()
		if int(auser.GetUserId()) == viewId {
			// 名前を取り出す
			dataTemp["name"] = auser.GetUserName()
		}
	}

	// 読んでいる位置
	index := -1
	if len(queries["index"]) > 0 {
		indexTmp, _ := strconv.ParseInt(queries["index"][0], 10, 0)
		index = int(indexTmp)
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
		network.RedirectLogonTop(w, r, "", "存在しないキーワードです")
		return
	}

	// キーワードに対応した記事番号を取り出す
	akeyword, err := network.GetKeywordDetail(client, keywordId)

	// 記事番号をマップにまとめておく
	mapItemId := make(map[int]int, len(akeyword.GetKeywordItemSet()))
	for _, keywordItem := range akeyword.GetKeywordItemSet() {
		mapItemId[int(keywordItem.GetItemId())] = 1
	}

	// 記事情報を取り出す(この中にキーワードに対応する記事がある)
	itemInfo, err := network.GetItemInfoByUserId(client, w, r, viewId)

	if err != nil {
		network.RedirectLogonTop(w, r, "", err.Error())
		return
	}

	// 可視化用に調整する
    mapItem := make(map[int]*aitem3)
	cntItem := 0
	for _, aitem := range itemInfo.GetItemList() {

		itemId := int(aitem.GetItemId())
		_, ok := mapItemId[itemId]
		if !ok {
			continue
		}
		cntItem++
	}
	if index < 0 {
		index = 0
	}
	if index < 0 {
		index = 0
	}
	dataTemp["before"] = 0
	dataTemp["after"] = 0

	i := -1
	for _, aitem := range itemInfo.GetItemList() {

		itemId := int(aitem.GetItemId())
		_, ok := mapItemId[itemId]
		if !ok {
			continue
		}
		i++
		if i < index {
			dataTemp["before"] = 1
		} else if i >= index && i <= index+pagesize {
			strHeader := *aitem.HeaderString + "\n\n\n\n\n\n"
			newstr := strings.Join(strings.Split(strHeader, "\n")[0:5], "\n")
			//aitem.HeaderString = &newstr
            mapItem[i-index] = new(aitem3)
			mapItem[i-index].HeaderString = &newstr
			mapItem[i-index].ItemId = aitem.GetItemId()

            //mapItem[i-index].Created = strconv.FormatInt(aitem.GetCreated()*100-time.Date(2000,1,1,0,0,0,0,time.UTC).UnixNano(),10)
            mapItem[i-index].Created = DateTimeString(aitem.GetCreated())
            mapItem[i-index].LastModified = DateTimeString( aitem.GetLastModified())
		} else {
			dataTemp["after"] = 1
		}
	}
	dataTemp["index"] = index

	dataTemp["itemlist"] = mapItem
	dataTemp["pagesize"] = pagesize

	// レンダリング
	template.Execute("list_info_by_keyword", w, dataTemp)
}
