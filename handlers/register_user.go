package handlers

import (
	"code.google.com/p/go.image/bmp"
	"github.com/nfnt/resize"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

// RegisterUserへのリダイレクト1
func RedirectRegisterUser(w http.ResponseWriter, r *http.Request, message string) {

	http.Redirect(w, r, "/register_user?message="+message, http.StatusMovedPermanently)
}

// 記事の書き込み
func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {

	// POSTされたものかのチェック
	if r.Method != "POST" {

		// ページの表示用
		dataTemp := make(map[string]interface{})

		queries := r.URL.Query()
		if len(queries["message"]) > 0 {
			dataTemp["message"] = queries["message"][0]
		}

		// レンダリング
		template.Execute("register_user", w, dataTemp)
	} else {

		r.ParseMultipartForm(1024 * 1024 * 10)
		mapForm := r.MultipartForm.Value
		name := mapForm["username"][0]
		email := mapForm["email"][0]
		password := mapForm["password"][0]

		if len(name) == 0 {
			RedirectRegisterUser(w, r, "キャラクタ名を入力してください")
			return
		}
		if len(email) == 0 {
			RedirectRegisterUser(w, r, "EMailを入力してください")
			return
		}
		if len(password) == 0 {
			RedirectRegisterUser(w, r, "passwordを入力してください")
			return
		}

		f1, handler, err := r.FormFile("image")
		if err != nil {
			RedirectRegisterUser(w, r, "イメージファイルを指定してください")
			return
		}

		filepath1 := path.Join(os.TempDir(), handler.Filename)
		f2, err := os.Create(filepath1)
		if err != nil {
			RedirectRegisterUser(w, r, "Internal error : テンポラリファイルが作れませんでした")
			return
		}
		io.Copy(f2, f1)
		f2.Close()

		f2, err = os.Open(filepath1)
		img, _, err := image.Decode(f2)
		f1.Close()
		if err != nil {
			f2.Close()
			RedirectRegisterUser(w, r, "扱えない画像書式です : "+err.Error())
			return
		}
		img2 := resize.Resize(64, 64, img, resize.Lanczos3)

		f2.Close()
		f2, err = os.Create(filepath1 + ".bmp")
		err = bmp.Encode(f2, img2)

		f2.Close()
		imgData, err := ioutil.ReadFile(filepath1 + ".bmp")
		clientAddress := r.RemoteAddr

		err = network.AddUser(w, r, email, password, name, imgData, clientAddress)
		if err != nil {
			RedirectRegisterUser(w, r, err.Error())
			return
		}
		network.RedirectIndex(w, r, "", "ユーザ登録しました. ログインしてください")
	}
}
