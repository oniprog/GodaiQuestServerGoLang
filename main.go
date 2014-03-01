/*
  Godai Quest Http Server
*/
package main

import (
	"fmt"
	"github.com/eknkc/amber"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/oniprog/GodaiQuestServerGoLang/handlers"
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"log"
	"net/http"
	"os"
	"path"
)

var amberFolder = "./amber"
var amberOptions = amber.Options{PrettyPrint: false, LineNumbers: false}

var store = sessions.NewCookieStore([]byte("godaiquest-secret"))

// ファイルを返すだけのハンドラ
func fileHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("%s\n", r.URL.Path[1:])
	http.ServeFile(w, r, path.Join("public", r.URL.Path[1:]))
}

func main() {

	// テンプレートの準備
	err := template.Prepare(amberFolder, amberOptions)
	if err != nil {
		fmt.Printf("template compile error\n")
		os.Exit(1)
	}

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

	http.Handle("/", r)

	fmt.Printf("Server start : port 3001\n")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
