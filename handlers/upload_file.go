package handlers

import (
	"fmt"
	"os"
	"io"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"net/http"
	"errors"
	"path"
	"strconv"
)

//  情報を読む
func UploadFileHandler(w http.ResponseWriter, r *http.Request) {

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
		err = errors.New("書き込む情報の指定がありません")
		network.RedirectInfoTop(w, r, "", err.Error())
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
	for _, aitemdic := range itemInfo.GetAitemDic() {

		aitem := aitemdic.GetAitem()
		itemId := int(aitem.GetItemId())
		if itemId == infoId {
			find = true
			break
		}
	}
	if !find {

		network.RedirectInfoTop(w, r, "", "自分の情報に対してのみアップロードできます")
		return
	}

	downloadDir := path.Join(network.DownloadRoot, strconv.FormatUint(uint64(infoId), 10))
	os.MkdirAll(downloadDir, 0777)

	r.ParseMultipartForm(1024 * 1024 * 100)
	for i := 0; i < 100; i++ {
		filename := fmt.Sprintf("b%d", i)
		file, handler, err := r.FormFile(filename)
		if err != nil {
			continue
		}

		filepath1 := path.Join(downloadDir, handler.Filename)
		fmt.Printf("Upload file : %s\n", filepath1)

		f, err := os.Create(filepath1)
		if err != nil {
			fmt.Printf("Error : %s\n", err.Error())
			file.Close()
			continue
		}
		io.Copy(f, file)
		file.Close()
		f.Close()
	}

	ReadInfoHandler(w, r)
}

