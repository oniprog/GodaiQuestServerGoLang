/*
  Godai Quest Http Server
*/
package main

import (
	"encoding/json"
	"fmt"
	"github.com/eknkc/amber"
	"github.com/gorilla/mux"
	"github.com/oniprog/GodaiQuestServerGoLang/handlers"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"log"
	"net/http"
	"os"
	"path"
)

const amberFolder = "./amber"

var amberOptions = amber.Options{PrettyPrint: false, LineNumbers: false}

const downloadRoot = "public/download"

type Configuration struct {
	ServerAddr string
	Secret     string
}

// 設定ファイルの読み込み
func readConfig() (*Configuration, error) {

	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(file)
	configuration := &Configuration{}
	decoder.Decode(&configuration)

	return configuration, nil
}

// ファイルを返すだけのハンドラ
func fileHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("%s\n", r.URL.Path[1:])
	http.ServeFile(w, r, path.Join("public", r.URL.Path[1:]))
}

func makeNewRoute() {

	r := mux.NewRouter()

	// ./public以下を静的コンテンツの置き場所にする
	fileServer := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fileServer))

	// favicon.icoの処理
	r.HandleFunc("/favicon.ico", fileHandler)

	// '/'の処理
	r.HandleFunc("/", handlers.IndexHandler)
	r.HandleFunc("/index", handlers.IndexHandler)

	// 'login'の処理
	r.HandleFunc("/login", handlers.LoginHandler)

	// ユーザリスト
	r.HandleFunc("/list_user", handlers.ListUserHandler)

	// 情報リスト
	r.HandleFunc("/list_info", handlers.ListInfoHandler)
	r.HandleFunc("/list_info_all", handlers.ListInfoAllHandler)

	// 情報
	r.HandleFunc("/read_info", handlers.ReadInfoHandler)
	r.HandleFunc("/modify_info", handlers.ModifyInfoHandler)
	r.HandleFunc("/write_info", handlers.WriteInfoHandler)

	// 情報内の記事
	r.HandleFunc("/write_article", handlers.WriteArticleHandler)
	r.HandleFunc("/delete_last_article", handlers.DeleteLastAritcleHandler)

	// ファイル
	r.HandleFunc("/upload_file", handlers.UploadFileHandler)
	r.HandleFunc("/delete_file", handlers.DeleteFileHandler)

	// ユーザ登録
	r.HandleFunc("/register_user", handlers.RegisterUserHandler)

	// キーワード
	r.HandleFunc("/attach_keyword", handlers.AttachKeywordHandler)
	r.HandleFunc("/list_info_by_keyword", handlers.ListInfoByKeywordHandler)
	r.HandleFunc("/manage_keyword", handlers.ManageKeywordHandler)

	// ログアウト
	r.HandleFunc("/logout", handlers.LogoutHandler)

	http.Handle("/", r)
}

func main() {

	//
	config, err := readConfig()
	if err != nil {
		fmt.Printf("Configuration read fail : " + err.Error())
		os.Exit(1)
	}

	// テンプレートの準備
	err = template.Prepare(amberFolder, amberOptions)
	if err != nil {
		fmt.Printf("template compile error\n")
		os.Exit(1)
	}

	// ネットワークの初期化
	err = network.Prepare(config.ServerAddr, downloadRoot)
	if err != nil {
		fmt.Printf("network initialization error\n")
		os.Exit(1)
	}

	// セッションの準備
	err = sessions.Prepare(config.Secret)
	if err != nil {
		fmt.Printf("sessions error\n")
		os.Exit(1)
	}

	makeNewRoute()

	fmt.Printf("Server start : port 3001\n")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
