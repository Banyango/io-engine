// +build !js,wasm

package server

import (
	"encoding/json"
	"fmt"
	"github.com/pion/webrtc"
)

func DebugPeerConnection(connection *webrtc.PeerConnection) {
	bytes, _ := json.Marshal(connection.GetStats())
	fmt.Println(string(bytes))
}
