package network

import (
    "crypto/sha512"
    "fmt"
    "net/http"
    "code.google.com/p/goprotobuf/proto"
    "github.com/oniprog/GodaiQuestServerGoLang/godaiquest"
)

// 同期をとるためのオブジェクト
var lock = make(chan int, 1)

//
func RedirectIndex( w http.ResponseWriter, r *http.Request, email string, message string ) {

    if len(message) > 0 {
        http.Redirect(w, r, "/index?message="+message+"&email="+email, http.StatusSeeOther)
    } else {
        http.Redirect(w, r, "/index?email="+email, http.StatusSeeOther)
    }
}


// ログインを試みる
func TryLogon(w http.ResponseWriter, r *http.Request) *Client {

    lock <- 1
    defer func() {<- lock}()

    //
    email := r.PostFormValue("email")
    password := r.PostFormValue("password")

    conn, err := Connect()
    if err != nil {
        fmt.Printf("%s\n", err.Error() )
        return nil
    }
    client := NewClient(conn)

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
    client.WriteDword( COM_TryLogon )
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
          ClientVersion: proto.Uint32(CLIENT_VERSION),
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


// すべてのユーザの情報を得る
func GetAllUesrInfo( client *Client ) {


}