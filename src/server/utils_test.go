package server

import (
	"bytes"
	"encoding/gob"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncode(t *testing.T) {
	netData := NetworkData{NetworkId:0, OwnerId:0}

	netData.Data = map[int][]byte{}

	var data struct{
		X int
		Y int
	}

	data.X = 33
	data.Y = 23

	EncodeNetworkDataBytes(&netData, 0, data)

	var result struct{
		X int
		Y int
	}

	reader := bytes.NewReader(netData.Data[0])
	enc := gob.NewDecoder(reader)
	err := enc.Decode(&result)
	if err != nil {
		assert.Fail(t, "error decoding")
	}

	DecodeNetworkDataBytes(&netData, 0, &result)

	assert.Equal(t, result.X, data.X)
	assert.Equal(t, result.Y, data.Y)
}
