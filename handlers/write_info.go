package handlers

import (
	"code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/oniprog/GodaiQuestServerGoLang/dungeon"
	"github.com/oniprog/GodaiQuestServerGoLang/godaiquest"
	"github.com/oniprog/GodaiQuestServerGoLang/network"
	"github.com/oniprog/GodaiQuestServerGoLang/sessions"
	"github.com/oniprog/GodaiQuestServerGoLang/template"
	"net/http"
)

// 記事の書き込み
func WriteInfoHandler(w http.ResponseWriter, r *http.Request) {

	// ログインチェック
	client, err := sessions.GetClient(w, r)
	if err != nil {
		network.RedirectIndex(w, r, "", err.Error())
		return
	}

	// ページの表示用
	dataTemp := make(map[string]interface{})

	// ダンジョン1階の情報を得る
	dungeon1, err := network.GetDugeon(client, client.UserId, 0 /*level*/)
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	// オブジェクトの情報取得
	objectAttrInfo, err := network.GetObjectAttrInfo(client)
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	// ダンジョンの空きスペースをチェックする
	dungeon1Maze := dungeon.ExtractMaze(dungeon1)
	cntSpace := dungeon.CountSpace(dungeon1Maze, dungeon.MakeObjIdToItemIdMap(objectAttrInfo) )

	if cntSpace == 0 {
		network.RedirectInfoTop(w, r, "", "ダンジョンを広げてください。スペースがありません")
		return
	}
	dataTemp["rest_item_cnt"] = fmt.Sprintf("%d", cntSpace)

	// POSTされたものかのチェック
	if r.Method != "POST" {

		// レンダリング
		template.Execute("write_info", w, dataTemp)
		return
	}


	// ブロックイメージの取得
	dungeonImagesInfo, err := network.GetDungeonImageBlock(client)
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	// タイル情報の取得
	tileInfo, err := network.GetTileList(client)
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	// 大陸
	dungeon0, err := network.GetDugeon(client, 0 /*大陸*/, 0 /*level*/)
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}
	// 大陸の自分の領地のある範囲
	islandGroundInfo, err := network.GetIslandGroundInfoByUser(client, client.UserId)
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}
	// 大陸に入り口を設置する
	dungeon.SetDungeonEntrance(dungeon0, islandGroundInfo, dungeon.MakeObjToAttrMap(objectAttrInfo), dungeon.MakeObjToImageId(tileInfo))

	islandProto := &godaiquest.SetDungeon{
		UserId:        proto.Int32(0),
		DungeonNumber: proto.Int32(0),
		DungeonInfo:   dungeon0,
		Images:        dungeonImagesInfo,
		ObjectInfo:    objectAttrInfo,
		TileInfo:      tileInfo,
	}
	err = network.SetDungeon(client, islandProto)

	// 書き込み内容
	newText := r.PostFormValue("inputtext")

	// ダンジョン内に情報を配置する
	imagepair := dungeon.GetSomeItemImagePair(dungeonImagesInfo, 0)
	newItem, err := network.CreateAItem(client, objectAttrInfo, imagepair, newText)
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}
	dungeonProto := &godaiquest.SetDungeon{
		UserId:        proto.Int32(int32(client.UserId)),
		DungeonNumber: proto.Int32(0),
		DungeonInfo:   dungeon1,
		Images:        dungeonImagesInfo,
		ObjectInfo:    objectAttrInfo, // 更新されている
		TileInfo:      tileInfo,
	}
	err = network.SetDungeon(client, dungeonProto)
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	// オブジェクトの再情報取得
	objectAttrInfo, err = network.GetObjectAttrInfo(client)
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}
	// ブロックイメージの再取得
	dungeonImagesInfo, err = network.GetDungeonImageBlock(client)
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	// タイル情報の再取得
	tileInfo, err = network.GetTileList(client)
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	// ダンジョン内に配置する
	err = dungeon.SetItemToEmtpyArea(dungeon1, dungeon.MakeItemIdToObjIdMap(objectAttrInfo), dungeon.MakeObjIdToItemIdMap(objectAttrInfo), newItem)
	if err != nil {
		network.RedirectInfoTop(w, r, "", err.Error())
		return
	}

	// 最終書き込み
	dungeonProto = &godaiquest.SetDungeon{
		UserId:        proto.Int32(int32(client.UserId)),
		DungeonNumber: proto.Int32(0),
		DungeonInfo:   dungeon1,
		Images:        dungeonImagesInfo,
		ObjectInfo:    objectAttrInfo,
		TileInfo:      tileInfo,
	}
	err = network.SetDungeon(client, dungeonProto)

	ListInfoAllHandler(w, r)
}
