package network

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
)

var Host string = ""
var DownloadRoot = ""

const BUFFER_SIZE = 1024 * 1024
const CLIENT_VERSION = 2014021819

// 接続管理
type Client struct {
	Conn       *net.TCPConn
	Buffer     *bytes.Buffer
	ReadedSize int // 読み込み済みのバイト数(バッファの切り詰めに使用する)
	DataBuf    []byte
	UserId     int
}

// 接続の準備をする
func Prepare(host string, downloadRoot string) error {
	Host = host
	DownloadRoot = downloadRoot

	os.Mkdir(DownloadRoot, 0700)

	return nil
}

// 接続をする。
func Connect() (*net.TCPConn, error) {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", Host)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Clientの新規作成
func NewClient(conn *net.TCPConn) *Client {

	buf := make([]byte, 0, BUFFER_SIZE)
	return &Client{conn, bytes.NewBuffer(buf), 0, make([]byte, 1024), 0}
}

// dword書き込み
func (this *Client) WriteDword(dword int) error {

	data := []byte{
		byte(dword >> 24 & 0xff),
		byte(dword >> 16 & 0xff),
		byte(dword >> 8 & 0xff),
		byte(dword & 0xFF),
	}
	n, err := this.Conn.Write(data)
	if n != 4 {
		return errors.New("書き込めませんでした")
	}
	return err
}

// dword書き込み
func (this *Client) WriteDwordRev(dword int) error {

	data := []byte{
		byte(dword & 0xFF),
		byte(dword >> 8 & 0xff),
		byte(dword >> 16 & 0xff),
		byte(dword >> 24 & 0xff),
	}
	n, err := this.Conn.Write(data)
	if n != 4 {
		return errors.New("書き込めませんでした")
	}
	return err
}

// bytes書き込み
func (this *Client) WriteProtoData(data *[]byte) error {

	err := this.WriteDwordRev(len(*data))
	_, err = this.Conn.Write(*data)
	return err
}

// bytes読み込み
func (this *Client) ReadProtoData(err error) (*[]byte, error) {

	if err != nil {
		return nil, err
	}

	size, err := this.ReadDwordRev(err)
	if err != nil {
		return nil, err
	}
	err = this.EnsureReadByte(size)
	ret := make([]byte, size)
	this.Buffer.Read(ret)
	return &ret, nil
}

// dword読み込み
func (this *Client) ReadDword(err error) (int, error) {

	if err != nil {
		return 0, err
	}

	err = this.EnsureReadByte(4)
	if err != nil {
		return 0, err
	}

	var i32 int32
	binary.Read(this.Buffer, binary.BigEndian, &i32)
	this.ReadedSize += 4
	this.Refresh()

	return int(i32), nil
}

// dword読み込み
func (this *Client) ReadDwordRev(err error) (int, error) {

	if err != nil {
		return 0, err
	}

	err = this.EnsureReadByte(4)
	if err != nil {
		return 0, err
	}

	var i32 int32
	binary.Read(this.Buffer, binary.LittleEndian, &i32)
	this.ReadedSize += 4
	this.Refresh()

	return int(i32), nil
}

// 1バイト読み込む
func (this *Client) ReadByte(err error) (byte, error) {

	if err != nil {
		return 0, err
	}
	err = this.EnsureReadByte(1)
	if err != nil {
		return 0, err
	}

	ret, err := this.Buffer.ReadByte()
	this.ReadedSize += 1
	this.Refresh()

	return ret, err
}

// 1バイト書き込む
func (this *Client) WriteByte(data byte) error {

	abyte := make([]byte, 1)
	abyte[0] = data
	_, err := this.Conn.Write(abyte)
	return err
}

// 長さを書き込む
func (this *Client) WriteLength(len int) {

	if len < 0x10 {
		this.WriteByte(byte(len))
	} else if len <= 0xFFF {
		this.WriteByte(byte(len>>8 | 0x10))
		this.WriteByte(byte(len & 0xff))
	} else if len <= 0xFFFFF {
		this.WriteByte(byte(len>>16 | 0x20))
		this.WriteByte(byte(len >> 8 & 0xff))
		this.WriteByte(byte(len & 0xff))
	} else {
		this.WriteByte(byte(len>>24 | 0x30))
		this.WriteByte(byte(len >> 16 & 0xff))
		this.WriteByte(byte(len >> 8 & 0xff))
		this.WriteByte(byte(len & 0xff))
	}
}

// 長さを読み込む
func (this *Client) ReadLength(err error) (int, error) {

	if err != nil {
		return 0, err
	}
	lenbyte, err := this.ReadByte(err)
	if lenbyte < 0x10 {
		return int(lenbyte), err
	}

	readlength := (lenbyte >> 4) & 0x0f
	data := int(lenbyte & 0x0f)
	for i := byte(0); i < readlength; i++ {
		adata, err := this.ReadByte(err)
		if err != nil {
			return 0, err
		}
		data = data*0x100 + int(adata)
	}
	return data, err
}

// 読み込めることを保証する
func (this *Client) EnsureReadByte(ensurebyte int) error {

	for this.Buffer.Len() < ensurebyte {
		n, err := this.Conn.Read(this.DataBuf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error : %s", err)
			} else {
				continue
			}
			return err
		}
		if n == 0 {
			continue
		}
		this.Buffer.Write(this.DataBuf[:n])
	}
	return nil
}

// 文字列を書き込む
func (this *Client) WriteString(str string) error {

	data := StringToBinaryUTF16(str)
	this.WriteLength(len(data) + 1)

	_, err := this.Conn.Write(data)
	this.Conn.Write([]byte{0})
	return err
}

// 文字列を得る
func (this *Client) ReadString(err error) (string, error) {

	if err != nil {
		return "", err
	}

	strlength, err := this.ReadLength(nil)
	data := make([]byte, strlength)
	err = this.EnsureReadByte(strlength)
	this.Buffer.Read(data)

	ret := BinaryUTF16ToString(data)
	return ret, nil
}

// ファイル情報を送信する
func (this *Client) WriteFileInfo(listFiles []GodaiFileInfo, basedir string) {

	for _, fileinfo := range listFiles {

		this.WriteDword(fileinfo.Size)
		filerelpath := fileinfo.PartPath
		this.WriteString(filerelpath)
		fmt.Printf("file info : %s\n", filerelpath)
	}
	this.WriteDword(-1)
}

// バイナリを受信する
func (this *Client) ReadBinary(err error) ([]byte, error) {

	if err != nil {
		return nil, err
	}

	size, err := this.ReadLength(nil)

	err = this.EnsureReadByte(size)
	data := make([]byte, size)
	this.Buffer.Read(data)

	return data, err
}

// ファイルを受信する
func (this *Client) ReadFiles(basedir string, err error) error {

	if err != nil {
		return err
	}

	for {

		continueFlag, err := this.ReadByte(err)
		if err != nil {
			return err
		}
		if continueFlag == 0 {
			break
		}

		filename, err := this.ReadString(err)
		outputPath := path.Join(basedir, filepath.Clean(filename))
		os.MkdirAll( filepath.Dir(outputPath), 0777 )
		
		fmt.Printf("Receive file : %s -> %s\n", filename, outputPath)

		data, err := this.ReadBinary(err)
		gzipReader, _ := gzip.NewReader(bytes.NewReader(data))
		fo, err := os.Create(outputPath)
		if err != nil {
			fmt.Printf("Can't crate %s\n", filename )
			return err
		}
		io.Copy( fo, gzipReader )
		gzipReader.Close()
		fo.Close()
	}

	return err
}

// バッファを切り詰める
func (this *Client) Refresh() {
	if this.Buffer.Len() == 0 {
		this.Buffer.Reset()
	}
	if this.ReadedSize > BUFFER_SIZE/2 {
		this.Buffer = bytes.NewBuffer(this.Buffer.Bytes())
		this.ReadedSize = 0
	}
}

// 閉じる
func (this *Client) Close() {
	if this.Conn != nil {
		this.Conn.Close()
	}
	this.Conn = nil
}
