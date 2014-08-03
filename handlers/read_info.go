package handlers

import (
	//	"fmt"
	"errors"
	"github.com/oniprog/GodaiQuestServerGoLang/godaiquest"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"net/http"
	"strconv"
	"strings"
)

//  情報を読む
func ReadInfoHandler(w http.ResponseWriter, r *http.Request) {

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
		err = errors.New("見る情報の指定がありません")
	}
	// 見る情報
	infoId := 0
	if len(queries["info_id"]) > 0 {
		dataTemp["info_id"] = queries["info_id"][0]
		infoId64, _ := strconv.ParseInt(queries["info_id"][0], 10, 0)
		infoId = int(infoId64)
	} else {
		err = errors.New("見る情報の指定がありません")
	}
	if err != nil {
		network.RedirectIndex(w, r, "", err.Error())
		return
	}

	dataTemp["user_id"] = strconv.FormatUint(uint64(client.UserId), 10)

	// すべてのユーザ情報の読み込み
	/*	userInfo, err := network.GetAllUserInfo(client, w, r)
		if err != nil {
			network.RedirectIndex(w, r, "", err.Error())
			return
		}*/

	// 情報を取り出す
	itemInfo, err := network.GetItemInfoByUserId(client, w, r, viewId)
	if err != nil {
		network.RedirectIndex(w, r, "", err.Error())
		return
	}
	var curItem *godaiquest.AItem2
	for _, aitem := range itemInfo.GetItemList() {

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

	// アイテムの詳細情報を取り出す
	listFiles, err := network.GetAItem(client, w, r, infoId)

	// 表示用に整形
	mapFiles := make(map[string]interface{})
	for _, fileinfo := range listFiles {

		mapAFile := make(map[string]interface{})
		filepath := fileinfo.PartPath
		mapAFile["Path"] = filepath
		if strings.HasSuffix(filepath, ".jpg") || strings.HasSuffix(filepath, ".png") {
			mapAFile["ImagePath"] = filepath
		} else {
			mapAFile["ImagePath"] = ""
		}

		mapFiles[filepath] = mapAFile
	}
	dataTemp["listFiles"] = mapFiles

	// 読んだことにする
	network.ReadMarkAtArticle(client, infoId)

	// 記事の内容を読む
	articleContent, err := network.GetArticleString(client, infoId)

	dataTemp["article_content"] = articleContent

	if err != nil {
		network.RedirectIndex(w, r, "", err.Error())
		return
	}
	// レンダリング
	template.Execute("read_info", w, dataTemp)
}
