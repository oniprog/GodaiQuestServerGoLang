package template

import (
	"fmt"
	"github.com/eknkc/amber"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var mapTemplate = make(map[string]*template.Template)
var amberOptions = amber.Options{}

// テンプレートを実行する
func Execute(tempname string, w http.ResponseWriter, data interface{}) {
	mapTemplate[tempname].Execute(w, data)
}

// テンプレートを実行する
func ExecuteWithFunc(tempname string, w http.ResponseWriter, data interface{}, funcmap template.FuncMap) {
	mapTemplate[tempname].Funcs(funcmap)
	mapTemplate[tempname].Execute(w, data)
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

func Prepare(amberFolder string, opts amber.Options) error {

	amberOptions = opts

	// テンプレートの準備
	err := filepath.Walk(amberFolder, compileTemplate)
	if err != nil {
		fmt.Printf("compileTemplate failed: %v", err)
		os.Exit(1)
	}
	return err
}
