package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

// 工具类
// 整数转化为16进制
func IntToHex(num int64) []byte {
		buff := new(bytes.Buffer)
		err := binary.Write(buff,binary.LittleEndian,num)

		if err != nil {
			log.Panic(err)
		}
		return buff.Bytes()
}