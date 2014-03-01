/*
  Godai Quest Http Server
*/
package main

import (
	"fmt"
	"log"
	"os"
    "path"
	"github.com/eknkc/amber"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

var AmberFolder = "./amber"
var amberOptions = amber.Options{PrettyPrint: false, LineNumbers: false}

var store = sessions.NewCookieStore([]byte("godaiquest-secret"))
var mapTemplate = make(map[string]*template.Template)

// '/' の処理
func indexHandler(w http.ResponseWriter, r *http.Request) {

     data := make( map[string]string )

     queries := r.URL.Query()
     if len(queries["message"]) > 0 {
        data["message"] = queries["message"][0]
     }     
     email := r.Form["email"]
     password := r.Form["password"]

     if len(email) > 0 && len(password) > 0 {
        data["email"] = email[0]
        data["password"] = password[0]
     }     
	 mapTemplate["index"].Execute(w, data)
}

// ログインの処理
func loginHandler(w http.ResponseWriter, r *http.Request) {

     if r.Method != "POST" {
        http.Redirect( w, r, "/index", http.StatusMovedPermanently )
        return
     }

    http.Redirect( w, r, "/index?message=ログインしました", http.StatusSeeOther )
}

// テンプレートのコンパイル処理
func compileTemplate(path string, f os.FileInfo, err error) error {

	// フォルダは無視する
	if f.IsDir() {
		return nil
	}

	// .amberという拡張子を持つファイルだけを対象とする
	if !strings.HasSuffix(f.Name(), ".amber") {
		return nil
	}

	templateName := f.Name()[0 : len(f.Name())-6]
	var compiler = amber.New()
	compiler.Options = amberOptions

	err = compiler.ParseFile(path)
	if err == nil {
		template, err := compiler.Compile()
		if err == nil {
			mapTemplate[templateName] = template
			fmt.Printf("Compile template(%s) : %s\n", templateName, path)
		}
	}
	return err
}

// ファイルを返すだけのハンドラ
func fileHandler(w http.ResponseWriter, r *http.Request) {

     fmt.Printf("%s\n", r.URL.Path[1:])
     http.ServeFile(w, r, path.Join("public", r.URL.Path[1:]))
}

func main() {

	// テンプレートの準備
	err := filepath.Walk(AmberFolder, compileTemplate)
	if err != nil {
		fmt.Printf("compileTemplate failed: %v", err)
		os.Exit(1)
	}

	r := mux.NewRouter()

	// ./public以下を静的コンテンツの置き場所にする
	fileServer := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fileServer))

    // favicon.icoの処理
    r.HandleFunc("/favicon.ico", fileHandler);

	// '/'の処理
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/index", indexHandler)

    // 'login'の処理
    r.HandleFunc("/login", loginHandler)

	http.Handle("/", r)

    fmt.Printf("Server start : port 3001\n")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
