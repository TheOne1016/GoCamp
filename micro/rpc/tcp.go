package rpc

import (
	"encoding/binary"
	"net"
)

func ReadMsg(conn net.Conn) ([]byte, error) {
	//读取响应头
	lenBs := make([]byte, numOfLengthBytes)
	_, err := conn.Read(lenBs)
	if err != nil {
		return nil, err
	}

	//响应有多长
	length := binary.BigEndian.Uint64(lenBs)

	data := make([]byte, length)
	_, err = conn.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func EncodeMsg(data []byte) []byte {
	reqLen := len(data)
	res := make([]byte, reqLen+numOfLengthBytes)
	binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(reqLen))
	copy(res[numOfLengthBytes:], data)
	return res
}
