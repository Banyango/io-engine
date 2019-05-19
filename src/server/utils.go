package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func DecodeNetworkDataBytes(networkPacket *NetworkData, id int, buf interface{}) {
	reader := bytes.NewReader(networkPacket.Data[id])
	enc := gob.NewDecoder(reader)
	err := enc.Decode(&buf)
	if err != nil {
		fmt.Println("Error Decoding id:", id, " owner:", networkPacket.OwnerId)
	}
}

func EncodeNetworkDataBytes(networkPacket *NetworkData, id int, buf interface{}) {
	buffer := bytes.Buffer{}
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(buf)
	if err != nil {
		fmt.Println("Error encoding id:", id, " owner:", networkPacket.OwnerId, err)
	}
	networkPacket.Data[id] = buffer.Bytes()
}
