package handlers

import (
    "net/http"
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
)

// 処理
func ListUserHandler(w http.ResponseWriter, r *http.Request) {

     // ログインチェック
    client, err := sessions.GetClient( w, r )
    if err != nil {
        network.RedirectIndex( w, r, "", err.Error() )
        return
    }

    // すべてのユーザ情報の読み込み
    userInfo, err := network.GetAllUserInfo( client, w, r )
    if err != nil {
        network.RedirectIndex( w, r, "", err.Error() )
        return
    }

    // ページ用のデータ生成
	dataTemp := make(map[string]interface{})

    mapUsers := make(map[int]interface{})
    for _, auserdic := range userInfo.GetUesrDic() {
    
        auser := auserdic.GetAuser()
        mapUser := make(map[string]interface{})
        mapUser["UserId"] = auser.GetUserId()
        mapUser["EMail"] = auser.GetMailAddress()
        mapUser["UserName"] = auser.GetUserName()
        mapUser["UriImage"] = string(network.ConvURIImage( auser.GetUserImage() ))
        mapUsers[int(auser.GetUserId())] = mapUser
    }
    dataTemp["UserInfo"] = mapUsers


    /*
    maptest := make(map[string]string);
    maptest["list1"] = "abc"
    maptest["list2"] = "def"
    dataMap["body"] = maptest
*/
    //
	template.Execute("list_user", w, dataTemp)
}