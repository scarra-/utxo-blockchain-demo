package network

import (
	"bytes"
	"encoding/gob"
	"log"
)

func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	encoder := gob.NewEncoder(&buff)
	if err := encoder.Encode(data); err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func BytesToMsg(bytes []byte) string {
	var cmd []byte

	for _, b := range bytes {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}

	return string(cmd)
}
