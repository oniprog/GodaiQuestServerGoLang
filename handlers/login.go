/*
　ログイン処理を行う
*/
package handlers

import (
    "crypto/sha512"
    "fmt"
	"net/http"
    "code.google.com/p/goprotobuf/proto"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
    "github.com/oniprog/GodaiQuestServerGoLang/godaiquest"
)

//
func RedirectIndex( w http.ResponseWriter, r *http.Request, email string, message string ) {

    if len(message) > 0 {
        http.Redirect(w, r, "/index?message="+message+"&email="+email, http.StatusSeeOther)
    } else {
        http.Redirect(w, r, "/index?email="+email, http.StatusSeeOther)
    }
}


// 接続処理を行う 
func makeClient(w http.ResponseWriter, r *http.Request) *network.Client {

    //
    email := r.PostFormValue("email")
    password := r.PostFormValue("password")

    conn, err := network.Connect( )
    if err != nil {
        fmt.Printf("%s\n", err.Error() )
        return nil
    }
    client := network.NewClient(conn)

    // コマンド送信
    err = client.WriteDword( 1 ) // Version
    okcode, err := client.ReadDword( err )
    if err != nil {
        RedirectIndex( w, r, email, "サーバーとの接続に失敗しました("+ err.Error()+")")
        client.Close()
        return nil
    }
    if okcode != 1 {
        fmt.Printf("%d\n", okcode) 
        RedirectIndex( w, r, email, "サーバーとの接続に失敗しました")
        client.Close()
        return nil
    }

    // ログインコマンド送信
    client.WriteDword( network.COM_TryLogon )
    client.WriteDword( 1 ); // version

    if len(email) == 0 || len(password) == 0 {
        RedirectIndex( w, r, email, "全部指定してください");
        client.Close()
        return nil
    }

    hasher := sha512.New()
    hasher.Write( []byte( password ) )
    passwordHash := fmt.Sprintf("%x", hasher.Sum(nil))
       
    login := &godaiquest.Login {
          MailAddress: proto.String(email),
          Password: proto.String(passwordHash),
          ClientVersion: proto.Uint32(network.CLIENT_VERSION),
    }          
    data, err := proto.Marshal(login)
    client.WriteBytes( &data )
    
    okcode, err = client.ReadDword( err )
    switch(okcode) {
        case 1:
            // ログイン成功
            return client
        case 3:
            RedirectIndex( w, r, email, "パスワードが間違っています");
        case 4:
            RedirectIndex( w, r, email, "ユーザが存在しません");
        default:
            RedirectIndex( w, r, email, "内部エラーです");
    } 
    client.Close()
    return nil
}

// ログインの処理
func LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Redirect(w, r, "/index", http.StatusMovedPermanently)
		return
	}

    // 接続処理を行う
    client := makeClient(w,r)

    // 登録処理
    if client != nil {
        email := r.PostFormValue("email")
        sessions.SetNewClient( w, r, client, email )
        http.Redirect(w, r, "/list_user", http.StatusSeeOther)
    }
}
