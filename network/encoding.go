package network

import (
	"unicode/utf16"
	"unicode/utf8"
)

func BinaryUTF8ToString(data []byte) string {

	ret := ""
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		ret = ret + string(r)
		data = data[size:]
	}
	return ret
}

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
