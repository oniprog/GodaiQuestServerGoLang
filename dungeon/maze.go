package dungeon

import (
	"fmt"
	"math/rand"
	"github.com/oniprog/GodaiQuestServerGoLang/godaiquest"
	"errors"
)

var COMMAND_Nothing = 0;
var COMMAND_GoUp = 1;
var COMMAND_GoDown = 2;
var COMMAND_IntoDungeon = 3;
var COMMAND_GoOutDungeon = 4;

// ダンジョンの情報を扱いやすい形式に変える
func ExtractMaze( dungeon *godaiquest.DungeonInfo ) []uint32 {

	dungeonByte := dungeon.GetDungeon()

	retMaze := make([]uint32, len(dungeonByte)/4 )

	for i:=0; i<len(dungeonByte); i += 4 {

		retMaze[i/4] =
			uint32(dungeonByte[i+0] ) +
			uint32(dungeonByte[i+1] << 8) +
			uint32(dungeonByte[i+2] << 16 ) +
			uint32(dungeonByte[i+3] << 24)
	}

	return retMaze
}

// ItemIdからObjectIdへのマップ
type ItemIdToObjIdMap map[uint32]uint32
func MakeItemIdToObjIdMap( objectAttrInfo *godaiquest.ObjectAttrInfo ) ItemIdToObjIdMap {

	retMap := make(ItemIdToObjIdMap)
	for _, objAttrDic := range objectAttrInfo.GetObjectAttrDic() {

		objectAttr := objAttrDic.GetObjectAttr()
		retMap[ uint32(objectAttr.GetItemId()) ] =uint32(objectAttr.GetObjectId())
	}
	return retMap
}

// ObjectIdからItemIdへのマップ
type ObjIdToItemIdMap map[uint32]uint32
func MakeObjIdToItemIdMap( objectAttrInfo *godaiquest.ObjectAttrInfo ) ObjIdToItemIdMap {

	retMap := make(ObjIdToItemIdMap)
	for _, objAttrDic := range objectAttrInfo.GetObjectAttrDic() {

		objectAttr := objAttrDic.GetObjectAttr()
		retMap[ uint32(objectAttr.GetObjectId()) ] =uint32(objectAttr.GetItemId())
	}
	return retMap
}

// ObjectId から ObjectAttrへのマップ
type ObjectIdToAttrMap map[uint32]*godaiquest.ObjectAttr 

// アイテム(情報)かどうかを判定する
func IsItem( itemId uint32 ) bool {
	return itemId != 0
}

// マップを作成する
func MakeObjToAttrMap( objectAttrInfo *godaiquest.ObjectAttrInfo ) ObjectIdToAttrMap { 

	retMap := make(ObjectIdToAttrMap)
	for _, objAttrDic := range objectAttrInfo.GetObjectAttrDic() {

		objectAttr := objAttrDic.GetObjectAttr()
		retMap[ uint32(objectAttr.GetObjectId()) ] = objectAttr
	}
	return retMap
}

// イメージIdへのマップを作成する
type ObjectIdToImageIdMap map[uint32]uint32
func MakeObjToImageId(tilelist *godaiquest.TileInfo) ObjectIdToImageIdMap {

	ret := make(ObjectIdToImageIdMap)
	for _, tiledic := range tilelist.GetTileDic() {

		tile := tiledic.GetTile()
		tileId := tile.GetTileId()
		objId := uint32(tileId & 0xffffffff)
		imageId := uint32(tileId >> 32)
		ret[objId] = imageId
	}
	return ret
}

// ダンジョン内の空きスペースを数える
func CountSpace( maze [] uint32, mapObjIdToItemId ObjIdToItemIdMap) int {

	nCnt := 0
	for i:=0; i<len(maze); i+=2 {
		objId := maze[i+0]
		if !IsItem( mapObjIdToItemId[objId] ) {
			nCnt++
		}
	}

	return nCnt
}

// ダンジョンに配置する
func SetDungeonAt( dungeon *godaiquest.DungeonInfo, ix int, iy int, objId uint32, imageId uint32 ) {
	
	index := (ix+iy*int(dungeon.GetSizeX())) * 8

	dungeon.Dungeon[ index + 0 ] = byte(objId & 0xff)
	dungeon.Dungeon[ index + 1 ] = byte(objId >> 8 & 0xff)
	dungeon.Dungeon[ index + 2 ] = byte(objId >> 16 & 0xff)
	dungeon.Dungeon[ index + 3 ] = byte(objId >> 24 & 0xff)
	dungeon.Dungeon[ index + 4 ] = byte(imageId & 0xff)
	dungeon.Dungeon[ index + 5 ] = byte(imageId>> 8 & 0xff)
	dungeon.Dungeon[ index + 6 ] = byte(imageId>> 16 & 0xff)
	dungeon.Dungeon[ index + 7 ] = byte(imageId>> 24 & 0xff)
}

// ダンジョンの入り口を置く
func SetDungeonEntrance( dungeon *godaiquest.DungeonInfo, ground *godaiquest.IslandGround, mapObjIdToAttr ObjectIdToAttrMap, mapObjIdToImageId ObjectIdToImageIdMap ) {

	maze := ExtractMaze( dungeon )
	ix1 := int(ground.GetIx1())
	iy1 := int(ground.GetIy1())
	ix2 := int(ground.GetIx2())
	iy2 := int(ground.GetIy2())

	// 既に入り口がないかをチェックする
	for iy := iy1; iy <= iy2; iy++ {
		for ix :=ix1; ix<= ix2; ix++ {

			index := (ix+iy*int(dungeon.GetSizeX()))*2
			objId := maze[index+0]
			if int((mapObjIdToAttr)[objId].GetCommand()) == COMMAND_IntoDungeon {
				return
			}
		}
	}

	// 入り口のオブジェクトを得る
	var entranceObjId uint32 = 0
	for id, objattr := range mapObjIdToAttr {
		if int(objattr.GetCommand()) == COMMAND_IntoDungeon { 
			entranceObjId = id
			break
		}
	}
	if entranceObjId == 0 {
		fmt.Printf("入り口オブジェクトがありません")
		return
	}
	imageId := mapObjIdToImageId[uint32(entranceObjId)]

	// 強制書き込み
	iforcex := rand.Intn(ix2-ix1+1)+ ix1
	iforcey := rand.Intn(iy2-iy1+1)+ iy1

	SetDungeonAt( dungeon, iforcex, iforcey, entranceObjId, imageId )
}

// 情報のイメージを得る
func GetSomeItemImagePair( imageinfo *godaiquest.DungeonBlockImageInfo, index int ) *godaiquest.ImagePair {

	for _, imagePairDic := range imageinfo.GetImageDic() {

		imagePair := imagePairDic.GetImagepair()
		if imagePair.GetCanItemImage() {
			if index == 0 {
				return imagePair
			}
			index--
		}
	}

	return nil
}

// ダンジョン内に情報を配置する
func SetItemToEmtpyArea( dungeon *godaiquest.DungeonInfo, mapItemIdToObjId ItemIdToObjIdMap, mapObjIdToItemId ObjIdToItemIdMap, newItem *godaiquest.AItem) error {

	maze := ExtractMaze( dungeon )
	sizeX := int(dungeon.GetSizeX())
	sizeY := int(dungeon.GetSizeY())

	// 空きスペースを見つける
	ixf, iyf := 0, 0
	find := false
	for iy := 0; iy <sizeY ; iy++ {
		for ix :=0; ix<sizeX; ix++ {

			index := (ix+iy*int(dungeon.GetSizeX()))*2
			objId := maze[index+0]
			if !IsItem(mapObjIdToItemId[objId]) {
				find = true
				ixf, iyf = ix, iy
				break
			}
		}
		if find {
			break
		}
	}

	if !find {
		return errors.New("アイテムを置く場所がありませんでした")
	}

	SetDungeonAt( dungeon, ixf, iyf, mapItemIdToObjId[uint32(newItem.GetItemId())] , uint32(newItem.GetItemImageId() ))

	return nil
}


