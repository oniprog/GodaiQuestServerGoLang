package sessions

import (
	"errors"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"net/http"
	"time"
)

// タイムアウトするまでの待ち時間
const Timeout = 10 * 60 * 60

var store = sessions.NewCookieStore([]byte("i"))

var mapClient = make(map[int]*network.Client)
var mapClientAccess = make(map[int]int64)
var mapClientEmail = make(map[int]string)
var cntClient = 1

// 準備
func Prepare(secret string) error {
	store = sessions.NewCookieStore([]byte(secret))
	return nil
}

// 同じclientを削除する
func DeleteClientSameEmail(email string) {
	// 同じemailを持つ接続を切る
	for index, tmp_email := range mapClientEmail {
		if tmp_email == email {
			fmt.Printf("Delete client : %s(%d)\n", email, index)
			client := mapClient[index]
			delete(mapClient, index)
			delete(mapClientEmail, index)
			delete(mapClientAccess, index)
			client.Close()
			break
		}
	}
}

// 作成したclientを登録する
func SetNewClient(w http.ResponseWriter, r *http.Request, client *network.Client, email string) {

	// 登録作業
	mapClient[cntClient] = client
	mapClientAccess[cntClient] = time.Now().Unix()
	mapClientEmail[cntClient] = email

	session, _ := store.Get(r, "godaiquest")
	session.Values["ClientNumber"] = cntClient
	session.Save(r, w)

	fmt.Printf("Register client %d\n", cntClient)

	cntClient++
}

// 可能ならばセッションからデータを取り出す
func GetClient(w http.ResponseWriter, r *http.Request) (*network.Client, error) {

	OnMyIdle()

	session, err := store.Get(r, "godaiquest")
	if err != nil {
		return nil, err
	}

	clientNumber, ok := session.Values["ClientNumber"].(int)
	if !ok {
		return nil, errors.New("ログインしていません")
	}

	client, ok := mapClient[clientNumber]
	if !ok {
		return nil, errors.New("ログインしていません")
	}
	mapClientAccess[clientNumber] = time.Now().Unix()

	return client, nil
}

// ログアウト処理
func Logout(client *network.Client) {

	for index, elem := range mapClient {
		if elem == client {

			fmt.Printf("Logout %s(%d)\n", mapClientEmail[index], index)
			delete(mapClient, index)
			delete(mapClientAccess, index)
			delete(mapClientEmail, index)
			return
		}
	}
}

// 定期的に削除する
func OnMyIdle() {

	indexDelete := 0
	indexDeleteMax := 10
	listDelete := make([]int, 10)

	currentTime := time.Now().Unix()
	for index, time := range mapClientAccess {
		elapsedTime := currentTime - time
		if elapsedTime >= Timeout {
			listDelete[indexDelete] = index
			indexDelete++
			if indexDelete == indexDeleteMax {
				break
			}
		}
	}

	for i := 0; i < indexDelete; i++ {

		fmt.Printf("Delete client : %d\n", listDelete[i])
		client := mapClient[listDelete[i]]
		client.Close()
		delete(mapClient, listDelete[i])
		delete(mapClientAccess, listDelete[i])
		delete(mapClientEmail, listDelete[i])
	}
}
