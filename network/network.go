package network

import (
    "crypto/sha512"
    "errors"
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

    // ロックする
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
    client.WriteProtoData( &data )
    
    okcode, err = client.ReadDword( err )
    switch(okcode) {
        case 1:
            // ログイン成功
            userId, _:= client.ReadDword( err )
            client.UserId = userId
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
func GetAllUserInfo( client *Client, w http.ResponseWriter, r *http.Request ) (*godaiquest.UserInfo, error) {

    // ロックする
    lock <- 1
    defer func() {<- lock}()

    //
    client.WriteDword( COM_GetUserInfo )
    client.WriteDword( 1 )  // Version 

    var err error = nil
    okcode, err := client.ReadDword(err)
    if okcode != 1 {
       return nil, errors.New("ユーザ情報の取得に失敗しました")
    }

    data, err := client.ReadProtoData(err)
    newUserInfo := &godaiquest.UserInfo{}
    err = proto.Unmarshal( *data, newUserInfo )
    return newUserInfo, err
}