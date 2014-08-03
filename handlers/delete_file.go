package handlers

import (
	"errors"
	"fmt"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

// ファイルの削除
func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {

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

	// ユーザ
	viewId := client.UserId

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

	// ファイル名を得る
	filename := ""
	if len(queries["filename"]) > 0 {
		filename = queries["filename"][0]
	} else {
		network.RedirectInfoTop(w, r, "", "削除するファイル名の指定がありません")
		return
	}

	// アイテム情報を得る
	itemInfo, err := network.GetItemInfoByUserId(client, w, r, viewId)

	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	// 宛先が自分のものかをチェックする
	find := false
	for _, aitem := range itemInfo.GetItemList() {

		itemId := int(aitem.GetItemId())
		if itemId == infoId {
			find = true
			break
		}
	}
	if !find {

		network.RedirectInfoTop(w, r, "", "自分の情報に対してのみ削除できます")
		return
	}

	downloadDir := path.Join(network.DownloadRoot, strconv.FormatUint(uint64(infoId), 10))
	fmt.Printf("delete file : %s\n", filepath.Clean(filename))
	os.Remove(path.Join(downloadDir, filepath.Clean(filename)))

	ReadInfoHandler(w, r)
}
