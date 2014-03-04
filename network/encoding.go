package network

import (
	"unicode/utf16"
	"unicode/utf8"
)

// utf8バイナリを文字列に変換
func BinaryUTF8ToString(data []byte) string {

	ret := ""
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		ret = ret + string(r)
		data = data[size:]
	}
	return ret
}

// utf16バイナリを文字列に変換
func BinaryUTF16ToString(data []byte) string {

	data16 := make([]uint16, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		data16[i/2] = uint16(data[i+0]) + uint16(data[i+1])*0x100
	}
	runes := utf16.Decode(data16)

	ret := ""
	for _, rune := range runes {
		ret = ret + string(rune)
	}
	return ret
}

// 文字列をutf16バイナリに変換
func StringToBinaryUTF16(str string) []byte {

	runes := make([]rune, len(str))
	cnt := 0
	for _, rune := range str {
		runes[cnt] = rune
		cnt++
	}
	words := utf16.Encode(runes[:cnt])

	bytedata := make([]byte, len(words)*2)
	for i, word := range words {
		bytedata[i*2+0] = byte(word & 0xff)
		bytedata[i*2+1] = byte(word >> 8)
	}
	return bytedata
}
