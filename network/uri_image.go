package network

import (
    "encoding/base64"
)

// URIイメージに変換する
func ConvURIImage( data []byte ) string {

     return base64.StdEncoding.EncodeToString( data )
}