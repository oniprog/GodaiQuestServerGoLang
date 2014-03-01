/*
  Godai Quest Http Server
*/
package main

import (
	"fmt"
	"log"
	"os"
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

	mapTemplate["index"].Execute(w, nil)
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

	// '/'の処理
	r.HandleFunc("/", indexHandler)
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":3001", nil))
}
