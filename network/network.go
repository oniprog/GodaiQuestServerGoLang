package network

import (
	"code.google.com/p/goprotobuf/proto"
	"crypto/sha512"
	"errors"
	"fmt"
	"github.com/oniprog/GodaiQuestServerGoLang/godaiquest"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
)

// 同期をとるためのオブジェクト
var lock = make(chan int, 1)

// トップページへのリダイレクト
func RedirectIndex(w http.ResponseWriter, r *http.Request, email string, message string) {

	if len(message) > 0 {
		http.Redirect(w, r, "/index?message="+message+"&email="+email, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/index?email="+email, http.StatusSeeOther)
	}
}

// ログイントップへのリダイレクト
func RedirectLogonTop(w http.ResponseWriter, r *http.Request, email string, message string) {

	http.Redirect(w, r, "/list_user?message="+message, http.StatusSeeOther)
}

// ログイントップへのリダイレクト
func RedirectInfoTop(w http.ResponseWriter, r *http.Request, email string, message string) {

	http.Redirect(w, r, "/read_info?message="+message, http.StatusSeeOther)
}

// ログインを試みる
func TryLogon(w http.ResponseWriter, r *http.Request) *Client {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	//
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	conn, err := Connect()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return nil
	}
	client := NewClient(conn)

	// コマンド送信
	err = client.WriteDword(1) // Version
	okcode, err := client.ReadDword(err)
	if err != nil {
		RedirectIndex(w, r, email, "サーバーとの接続に失敗しました("+err.Error()+")")
		client.Close()
		return nil
	}
	if okcode != 1 {
		fmt.Printf("%d\n", okcode)
		RedirectIndex(w, r, email, "サーバーとの接続に失敗しました")
		client.Close()
		return nil
	}

	// ログインコマンド送信
	client.WriteDword(COM_TryLogon)
	client.WriteDword(1) // version

	if len(email) == 0 || len(password) == 0 {
		RedirectIndex(w, r, email, "全部指定してください")
		client.Close()
		return nil
	}

	hasher := sha512.New()
	hasher.Write([]byte(password))
	passwordHash := fmt.Sprintf("%x", hasher.Sum(nil))

	login := &godaiquest.Login{
		MailAddress:   proto.String(email),
		Password:      proto.String(passwordHash),
		ClientVersion: proto.Uint32(CLIENT_VERSION),
	}
	data, err := proto.Marshal(login)
	client.WriteProtoData(&data)

	okcode, err = client.ReadDword(err)
	switch okcode {
	case 1:
		// ログイン成功
		userId, _ := client.ReadDword(err)
		client.UserId = userId
		return client
	case 3:
		RedirectIndex(w, r, email, "パスワードが間違っています")
	case 4:
		RedirectIndex(w, r, email, "ユーザが存在しません")
	default:
		RedirectIndex(w, r, email, "内部エラーです")
	}
	client.Close()
	return nil
}

// すべてのユーザの情報を得る
func GetAllUserInfo(client *Client, w http.ResponseWriter, r *http.Request) (*godaiquest.UserInfo, error) {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	//
	client.WriteDword(COM_GetUserInfo)
	client.WriteDword(1) // Version

	var err error = nil
	okcode, err := client.ReadDword(err)
	if okcode != 1 {
		return nil, errors.New("ユーザ情報の取得に失敗しました")
	}

	data, err := client.ReadProtoData(err)
	newUserInfo := &godaiquest.UserInfo{}
	err = proto.Unmarshal(*data, newUserInfo)
	return newUserInfo, err
}

// 未取得アイテム数の取得
func GetUnpickedupItemInfo(client *Client, w http.ResponseWriter, r *http.Request, userId int, dungeonId int) ([]int, error) {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	const MAX_ITEM = 100
	retUserId := make([]int, MAX_ITEM)

	client.WriteDword(COM_GetUnpickedupItemInfo)
	client.WriteDword(1)         // version
	client.WriteDword(userId)    // 対象ユーザ
	client.WriteDword(dungeonId) // ダンジョンId

	okcode, err := client.ReadDword(nil)
	if okcode != 1 {
		return nil, errors.New("未取得アイテム情報の取得に失敗しました")
	}

	length, err := client.ReadLength(nil)
	retLength := 0
	for i := 0; i < length; i++ {

		Id, err := client.ReadDword(err)
		if err != nil {
			return nil, err
		}
		if retLength < MAX_ITEM {
			retUserId[retLength] = Id
			retLength++
		}
	}

	return retUserId[:retLength], err
}

// ユーザIdに対応するアイテム情報を得る
func GetItemInfoByUserId(client *Client, w http.ResponseWriter, r *http.Request, userId int) (*godaiquest.ItemInfo, error) {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	client.WriteDword(COM_GetItemInfoByUserId)
	client.WriteDword(1) // Version
	client.WriteDword(userId)

	okcode, err := client.ReadDword(nil)
	if okcode != 1 {
		return nil, errors.New("アイテム情報の取得に失敗しました")
	}

	data, err := client.ReadProtoData(err)
	retItemInfo := &godaiquest.ItemInfo{}
	err = proto.Unmarshal(*data, retItemInfo)
	return retItemInfo, err
}

// ファイル情報格納
type GodaiFileInfo struct {
	Name     string
	PartPath string
	Size     int
}

// ファイル情報を得る
func ReadDir(basedir string, dirpart string) []GodaiFileInfo {

	const MAX_FILE = 1000
	retArray := make([]GodaiFileInfo, MAX_FILE) // 1000個までしか扱わない
	listFiles, _ := ioutil.ReadDir(path.Join(basedir, dirpart))
	cnt := 0
	for _, info := range listFiles {

		if info.IsDir() {
			retTmp := ReadDir(basedir, path.Join(dirpart, info.Name()))
			for _, infoTmp := range retTmp {

				if cnt >= MAX_FILE {
					break
				}
				retArray[cnt] = infoTmp
				cnt++
			}
		} else {
			newFileInfo := GodaiFileInfo{info.Name(), path.Join(dirpart, info.Name()), int(info.Size())}
			if cnt >= MAX_FILE {
				break
			}
			retArray[cnt] = newFileInfo
			cnt++
		}
	}

	return retArray[:cnt]
}

// アイテム詳細情報を得る
func GetAItem(client *Client, w http.ResponseWriter, r *http.Request, infoId int) ([]GodaiFileInfo, error) {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	// ダウンロードフォルダ
	DownloadDir := path.Join(DownloadRoot, strconv.FormatUint(uint64(infoId), 10))

	// ファイル情報を得る
	listFiles := ReadDir(DownloadDir, "")

	//
	client.WriteDword(COM_GetAItem)
	client.WriteDword(2) // version
	client.WriteDword(infoId)
	client.WriteFileInfo(listFiles, DownloadDir)

	err := client.ReadFiles(DownloadDir, nil)

	//
	listRetFiles := ReadDir(DownloadDir, "")

	return listRetFiles, err
}

// アイテムを読んだことを記録する
func ReadMarkAtArticle(client *Client, infoId int) error {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	client.WriteDword(COM_ReadArticle)
	client.WriteDword(1) // version
	client.WriteDword(infoId)
	client.WriteDword(client.UserId)

	okcode, err := client.ReadDword(nil)
	if okcode != 1 {
		return errors.New("アイテムを読んだことを記録できませんでした")
	}
	if err != nil {
		return err
	}
	return nil
}

// 記事を読む
func GetArticleString(client *Client, infoId int) (string, error) {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	client.WriteDword(COM_GetArticleString)
	client.WriteDword(1) // version
	client.WriteDword(infoId)

	okcode, err := client.ReadDword(nil)
	if okcode != 1 {
		return "", errors.New("記事への書き込みを読むことができませんでした")
	}
	if err != nil {
		return "", err
	}

	ret, err := client.ReadString(err)
	return ret, err
}

// アイテム内の記事の書き込み
func SetItemArticle( client *Client, infoId int, articleId int, userId int, contents string ) error {

	// ロックする
	lock <- 1
	defer func() { <-lock }()
	
	client.WriteDword( COM_SetItemArticle )
	client.WriteDword( 1 ) // version

	itemArticle := &godaiquest.ItemArticle{
		ItemId : proto.Int32( int32(infoId) ),
		ArticleId : proto.Int32( int32(articleId) ),
		UserId : proto.Int32( int32(userId) ),
		Contents : proto.String( contents ),
		CretaeTime : proto.Int64(0),
	}
	data, err := proto.Marshal( itemArticle )
	client.WriteProtoData( &data )

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return err
	}
	if okcode != 1 {
		return errors.New("記事の書き込みに失敗しました")
	}

	return nil
}
