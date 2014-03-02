package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

var Host string = ""

const BUFFER_SIZE = 10240
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
func Prepare(host string) error {
	Host = host
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

// バッファを切り詰める
func (this *Client) Refresh() {
	if this.Buffer.Len() == 0 {
		this.Buffer.Reset()
	}
	if this.ReadedSize > BUFFER_SIZE/2 {
		this.Buffer = bytes.NewBuffer(this.Buffer.Bytes())
	}
}

// 閉じる
func (this *Client) Close() {
	if this.Conn != nil {
		this.Conn.Close()
	}
	this.Conn = nil
}
