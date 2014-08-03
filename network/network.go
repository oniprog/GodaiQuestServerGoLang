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

//サーバーとの接続
func ConnectServer(w http.ResponseWriter, r *http.Request, email string) *Client {

	// サーバとの接続
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

	return client
}

// ログインを試みる
func TryLogon(w http.ResponseWriter, r *http.Request) *Client {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	//
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	client := ConnectServer(w, r, email)
	if client == nil {
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

	okcode, err := client.ReadDword(err)
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
func GetItemInfoByUserId(client *Client, w http.ResponseWriter, r *http.Request, userId int) (*godaiquest.ItemInfo2, error) {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	client.WriteDword(COM_GetItemInfo2ByUserId)
	client.WriteDword(1) // Version
	client.WriteDword(userId)

	okcode, err := client.ReadDword(nil)
	if okcode != 1 {
		return nil, errors.New("アイテム情報の取得に失敗しました")
	}

	data, err := client.ReadProtoData(err)
	retItemInfo2 := &godaiquest.ItemInfo2{}
	err = proto.Unmarshal(*data, retItemInfo2)
	return retItemInfo2, err
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
func SetItemArticle(client *Client, infoId int, articleId int, userId int, contents string) error {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	client.WriteDword(COM_SetItemArticle)
	client.WriteDword(1) // version

	itemArticle := &godaiquest.ItemArticle{
		ItemId:     proto.Int32(int32(infoId)),
		ArticleId:  proto.Int32(int32(articleId)),
		UserId:     proto.Int32(int32(userId)),
		Contents:   proto.String(contents),
		CretaeTime: proto.Int64(0),
	}
	data, err := proto.Marshal(itemArticle)
	client.WriteProtoData(&data)

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return err
	}
	if okcode != 1 {
		return errors.New("記事の書き込みに失敗しました")
	}

	return nil
}

// アイテム内の記事の削除
func DeleteLastItemAritcle(client *Client, infoId int) error {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	//
	client.WriteDword(COM_DeleteLastItemArticle)
	client.WriteDword(1) // version
	client.WriteDword(infoId)

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return err
	}
	if okcode != 1 {
		return errors.New("情報内の記事の削除に失敗しました")
	}

	return nil
}

// 記事の内容書き込み
func ChangeAItem(client *Client, infoId int, imageId int, newText string) error {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	//
	client.WriteDword(COM_ChangeAItem)
	client.WriteDword(1) // version

	aitem := &godaiquest.AItem{
		ItemId:       proto.Int32(int32(infoId)),
		ItemImageId:  proto.Int32(int32(imageId)),
		HeaderString: proto.String(newText),
		BNew:         proto.Bool(false),
	}

	data, err := proto.Marshal(aitem)
	client.WriteProtoData(&data)

	okcode, err := client.ReadDword(err)
	if err != nil {
		return err
	}
	if okcode != 1 {
		return errors.New("情報の変更に失敗しました")
	}
	return nil
}

// オブジェクトの情報を得る
func GetObjectAttrInfo(client *Client) (*godaiquest.ObjectAttrInfo, error) {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	client.WriteDword(COM_GetObjectAttrInfo)
	client.WriteDword(1) // version

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return nil, err
	}
	if okcode != 1 {
		return nil, errors.New("オブジェクト情報の取得に失敗しました")
	}

	data, err := client.ReadProtoData(err)
	objectAttrInfo := &godaiquest.ObjectAttrInfo{}
	err = proto.Unmarshal(*data, objectAttrInfo)
	return objectAttrInfo, err
}

// イメージブロックを得る
func GetDungeonImageBlock(client *Client) (*godaiquest.DungeonBlockImageInfo, error) {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	client.WriteDword(COM_GetDungeonBlockImage)
	client.WriteDword(1) // version

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return nil, err
	}
	if okcode != 1 {
		return nil, errors.New("イメージ情報の取得に失敗しました")
	}

	data, err := client.ReadProtoData(err)
	dungeonImagesInfo := &godaiquest.DungeonBlockImageInfo{}
	err = proto.Unmarshal(*data, dungeonImagesInfo)
	return dungeonImagesInfo, err
}

// タイルイメージを得る
func GetTileList(client *Client) (*godaiquest.TileInfo, error) {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	//
	client.WriteDword(COM_GetTileList)
	client.WriteDword(1)

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return nil, err
	}
	if okcode != 1 {
		return nil, errors.New("タイル情報の取得に失敗しました")
	}

	data, err := client.ReadProtoData(err)
	tileList := &godaiquest.TileInfo{}
	err = proto.Unmarshal(*data, tileList)
	return tileList, err
}

// ダンジョンの情報を得る
func GetDugeon(client *Client, dungeonId int, level int) (*godaiquest.DungeonInfo, error) {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	//
	client.WriteDword(COM_GetDungeon)
	client.WriteDword(1)

	getDugeon := &godaiquest.GetDungeon{
		Id:            proto.Int32(int32(dungeonId)),
		DungeonNumber: proto.Int32(int32(level)),
	}
	data1, err := proto.Marshal(getDugeon)
	client.WriteProtoData(&data1)

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return nil, err
	}
	if okcode != 1 {
		return nil, errors.New("ダンジョンの取得に失敗しました")
	}

	data2, err := client.ReadProtoData(nil)
	dungeonInfo := &godaiquest.DungeonInfo{}
	err = proto.Unmarshal(*data2, dungeonInfo)

	return dungeonInfo, err
}

// 大陸の自分の領地のある範囲
func GetIslandGroundInfoByUser(client *Client, userId int) (*godaiquest.IslandGround, error) {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	//
	client.WriteDword(COM_GetIslandGroundInfo)
	client.WriteDword(1)

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return nil, err
	}
	if okcode != 1 {
		return nil, errors.New("大陸土地情報の取得に失敗しました")
	}

	data, err := client.ReadProtoData(err)
	islandGroundInfo := &godaiquest.IslandGroundInfo{}
	err = proto.Unmarshal(*data, islandGroundInfo)

	// 自分の土地を検索する
	for _, islandGround := range islandGroundInfo.GetGroundList() {

		if int(islandGround.GetUserId()) == userId {
			return islandGround, nil
		}
	}

	return nil, errors.New("自分の土地がありませんでした")
}

// ダンジョンを設定する
func SetDungeon(client *Client, newdungeon *godaiquest.SetDungeon) error {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	//
	client.WriteDword(COM_SetDungeon)
	client.WriteDword(1) // version

	data, err := proto.Marshal(newdungeon)
	client.WriteProtoData(&data)

	if err != nil {
		fmt.Printf("Internal Error : %s\n", err.Error())
	}

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return err
	}
	if okcode != 1 {
		return errors.New("ダンジョンの更新に失敗しました")
	}

	return nil
}

// アイテムを登録する
func SetAItem(client *Client, newItem *godaiquest.AItem, newImagePair *godaiquest.ImagePair) (*godaiquest.AItem, error) {

	// ロックする
	lock <- 1
	defer func() { <-lock }()

	//
	client.WriteDword(COM_SetAItem)
	client.WriteDword(1) // version

	data1, err := proto.Marshal(newItem)
	if err != nil {
		fmt.Printf("Internal Error : %s\n", err.Error())
	}
	data2, err := proto.Marshal(newImagePair)
	if err != nil {
		fmt.Printf("Internal Error : %s\n", err.Error())
	}
	client.WriteProtoData(&data1)
	client.WriteProtoData(&data2)

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return nil, err
	}
	if okcode != 1 {
		return nil, errors.New("アイテムの更新に失敗しました")
	}

	// ファイルの送信だけれども
	client.WriteByte(0)

	// アイテムを受信する
	data3, err := client.ReadProtoData(nil)
	aitem := &godaiquest.AItem{}
	err = proto.Unmarshal(*data3, aitem)
	return aitem, err
}

// アイテムを新規作成する
// objectAttrInfoが更新されるので注意。SetDungeonで書き込む必要がある
func CreateAItem(client *Client, objectAttrInfo *godaiquest.ObjectAttrInfo, imagePair *godaiquest.ImagePair, itemContents string) (*godaiquest.AItem, error) {

	newImagePair := &godaiquest.ImagePair{
		Number:       proto.Int32(imagePair.GetNumber()),
		Image:        imagePair.GetImage(),
		Name:         proto.String(imagePair.GetName()),
		Owner:        proto.Int32(imagePair.GetOwner()),
		Created:      proto.Int64(0),
		CanItemImage: proto.Bool(true),
		NewImage:     proto.Bool(false),
	}

	newAItem := &godaiquest.AItem{
		ItemId:       proto.Int32(0),
		ItemImageId:  proto.Int32(imagePair.GetNumber()),
		HeaderString: proto.String(itemContents),
		HeaderImage:  nil,
		BNew:         proto.Bool(true),
	}

	aitem, err := SetAItem(client, newAItem, newImagePair)
	if err != nil {
		return nil, err
	}

	// 新しいObjectIdを得る
	(*objectAttrInfo.NewId)++
	newId := *objectAttrInfo.NewId

	objectAttr := &godaiquest.ObjectAttr{
		ObjectId:   proto.Int32(newId),
		CanWalk:    proto.Bool(true),
		ItemId:     proto.Int32(aitem.GetItemId()),
		BNew:       proto.Bool(true),
		Command:    proto.Int32(int32(COMMAND_Nothing)),
		CommandSub: proto.Int32(0),
	}
	newObjectAttrDic := &godaiquest.ObjectAttrDic{
		Index:      proto.Int32(newId),
		ObjectAttr: objectAttr,
	}

	newdic := make([]*godaiquest.ObjectAttrDic, len(objectAttrInfo.ObjectAttrDic)+1)
	for i, dic := range objectAttrInfo.ObjectAttrDic {
		newdic[i] = dic
	}
	newdic[len(objectAttrInfo.ObjectAttrDic)] = newObjectAttrDic
	objectAttrInfo.ObjectAttrDic = newdic

	return aitem, err
}

// ユーザの追加
func AddUser(w http.ResponseWriter, r *http.Request, email string, password string, name string, imgbyte []byte, clientAddress string) error {

	client := ConnectServer(w, r, email)
	if client == nil {
		return nil
	}

	defer client.Close()

	hasher := sha512.New()
	hasher.Write([]byte(password))
	passwordHash := fmt.Sprintf("%x", hasher.Sum(nil))

	client.WriteDword(COM_AddUser)
	client.WriteDword(1)

	newUser := &godaiquest.AddUser{
		MailAddress:  proto.String(email),
		UserName:     proto.String(name),
		Password:     proto.String(passwordHash),
		UserFolder:   proto.String("c:\\tmp\\godaiquest"),
		ComputerName: proto.String(clientAddress),
		UserImage:    imgbyte,
	}
	data, err := proto.Marshal(newUser)
	client.WriteProtoData(&data)

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return err
	}
	switch okcode {

	case 2:
		return errors.New("既に同じユーザが存在します")
	case 1:
		return nil
	default:
		return errors.New("ユーザ登録エラーです")
	}
}

// キーワード一覧を得る
func ListKeyword(client *Client, userId int) (*godaiquest.KeywordUserInfo, error) {

	client.WriteDword(COM_ListKeyword)
	client.WriteDword(1)
	client.WriteDword(userId)

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return nil, err
	}
	if okcode != 1 {
		return nil, errors.New("キーワード一覧の取得に失敗しました")
	}
	data, err := client.ReadProtoData(err)
	newKeywordInfo := &godaiquest.KeywordUserInfo{}
	err = proto.Unmarshal(*data, newKeywordInfo)

	return newKeywordInfo, err
}

// キーワードを登録する
func RegisterKeyword(client *Client, keyword string) (int, error) {

	client.WriteDword(COM_RegisterKeyword)
	client.WriteDword(1)
	client.WriteString(keyword)
	client.WriteDword(10000) // priority

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return 0, err
	}
	if okcode != 1 {
		return 0, errors.New("キーワードの登録に失敗しました")
	}

	keywordId, err := client.ReadDword(err)
	return keywordId, err
}

// キーワードを記事にひもづける
func AttachKeyword(client *Client, infoId int, keywordId int, itemPriority int) error {

	client.WriteDword(COM_AttachKeyword)
	client.WriteDword(1)
	client.WriteDword(keywordId)
	client.WriteDword(infoId)
	client.WriteDword(itemPriority)

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return err
	}
	if okcode != 1 {
		return errors.New("アイテムへのキーワードの割り当てに失敗しました")
	}
	return nil
}

// 記事をキーワードから外す
func DetachKeyword(client *Client, infoId int, keywordId int) error {

	client.WriteDword(COM_DetachKeyword)
	client.WriteDword(1)
	client.WriteDword(keywordId)
	client.WriteDword(infoId)
	okcode, err := client.ReadDword(nil)
	if err != nil {
		return err
	}
	if okcode != 1 {
		return errors.New("キーワードからの記事を外すのに失敗しました")
	}
	return nil
}

// キーワードの詳細を得る（対応する記事一覧)
func GetKeywordDetail(client *Client, keywordId int) (*godaiquest.AKeyword, error) {

	client.WriteDword(COM_GetKeywordDetail)
	client.WriteDword(1)
	client.WriteDword(keywordId)

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return nil, err
	}
	if okcode != 1 {
		return nil, errors.New("キーワード詳細の取得に失敗しました")
	}

	data, err := client.ReadProtoData(err)
	newAKeyword := &godaiquest.AKeyword{}
	err = proto.Unmarshal(*data, newAKeyword)

	return newAKeyword, err
}

// キーワードを削除する
func DeleteKeyword(client *Client, keywordId int) error {

	client.WriteDword(COM_DeleteKeyword)
	client.WriteDword(1)
	client.WriteDword(keywordId)

	okcode, err := client.ReadDword(nil)
	if err != nil {
		return err
	}
	if okcode != 1 {
		return errors.New("キーワードの削除に失敗しました")
	}

	return nil
}
