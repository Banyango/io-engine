package server

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/Banyango/socker"
	"github.com/goburrow/dynamic"
	"github.com/pion/webrtc"
	. "io-engine-backend/src/shared"
	"sync"
)


/*
----------------------------------------------------------------------------------------------------------------
Network Server Global
----------------------------------------------------------------------------------------------------------------
*/

type ServerGlobal struct {
	BufferedChanges []EntityChange
	NetworkChanges []EntityChange
	mux            sync.Mutex
}

type EntityChange struct {
	NetworkId       uint16
	NetworkPrefabId byte
	PrefabId        byte
	Type            EntityChangeType
	Buffer          bool
}

func (self *ServerGlobal) Id() int {
	return ServerGlobalType
}

func (self *ServerGlobal) NetworkSpawn(networkId uint16, prefabId byte, networkPrefab byte, buffer bool) {
	self.mux.Lock()
	self.CreateChange(networkId,prefabId,networkPrefab,buffer, InstantiatedEntityChangeType)
	self.mux.Unlock()
}

func (self *ServerGlobal) NetworkDestroy(networkId uint16, prefabId byte, networkPrefab byte, buffer bool) {
	self.mux.Lock()
	self.CreateChange(networkId,prefabId,networkPrefab,buffer, DestroyedEntityChangeType)
	self.mux.Unlock()
}

func (self *ServerGlobal) CreateChange(networkId uint16, prefabId byte, networkPrefab byte, buffer bool, changeType EntityChangeType) EntityChange {
	change := EntityChange{NetworkId: networkId, PrefabId: prefabId, NetworkPrefabId: networkPrefab, Type:changeType , Buffer: buffer}

	self.NetworkChanges = append(self.NetworkChanges, change)

	if buffer {
		self.BufferedChanges = append(self.BufferedChanges, change)
	}

	return change
}

func (self *ServerGlobal) CreateGlobal(world *World) {

}

func (self *ServerGlobal) Clear() {
	self.NetworkChanges = self.NetworkChanges[:0]
}

/*
----------------------------------------------------------------------------------------------------------------
Network Server System
----------------------------------------------------------------------------------------------------------------
*/

type ConnectionHandlerSystem struct {
	Connections          Storage
	MasterDataChannel    *webrtc.DataChannel
	MasterPeerConnection *webrtc.PeerConnection
}

func (self *ConnectionHandlerSystem) Init() {
	self.Connections = NewStorage()

	dynamic.Register("ServerGlobal", func() interface{} {
		return &ServerGlobal{}
	})
}

func (self *ConnectionHandlerSystem) AddToStorage(entity Entity) {
	for k := range entity.Components {
		component := entity.Components[k].(Component)

		if component.Id() == int(NetworkConnectionComponentType) {
			self.Connections.Components[entity.Id] = &component
		}
	}
}

func (*ConnectionHandlerSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{NetworkConnectionComponentType}
}

func (self *ConnectionHandlerSystem) UpdateSystem(delta float64, world *World) {

	//global := world.Globals[ServerGlobalType].(*ServerGlobal)
	//
	//// build the network packet.
	//
	//networkPacket := BufferedEntityPacket{
	//	Time:    time.Now().Unix(),
	//	Updates: global.NetworkChanges,
	//}
	//
	//var buf bytes.Buffer
	//enc := gob.NewEncoder(&buf)
	//
	//if err := enc.Encode(networkPacket); err != nil {
	//	fmt.Printf("Encoding Failed")
	//	return
	//}
	//
	//global.Clear()
}

/*
----------------------------------------------------------------------------------------------------------------
Network Shared structs
----------------------------------------------------------------------------------------------------------------
*/

// this connection must be a goroutine use a goroutine instead of component thing
type ClientInputPacket struct {
	PlayerId uint16
	Input    byte
}

type BufferedEntityPacket struct {
	Time    int64
	Updates []EntityChange
}

type ServerConnectionHandshakePacket struct {
	PlayerId uint16
	// remove this
	BufferedChanges []EntityChange
	Offer webrtc.SessionDescription
}

type EntityChangeType byte

const (
	InstantiatedEntityChangeType EntityChangeType = iota
	DestroyedEntityChangeType
	StateUpdatedEntityChangeType
)

type NetworkConnectionComponent struct {
	PlayerId uint16
	// websocket handler.
	connHandler socker.SockerClientConnection

	// Udp - unreliable
	PeerConnection *webrtc.PeerConnection
	DataChannel    *webrtc.DataChannel
	Offer          webrtc.SessionDescription

}

func (*NetworkConnectionComponent) Id() int {
	return int(NetworkConnectionComponentType)
}

func (self *NetworkConnectionComponent) CreateComponent() {

}

func (*NetworkConnectionComponent) DestroyComponent() {

}


func (self *NetworkConnectionComponent) Handshake(playerId uint16, buffer []EntityChange) {

	networkPacket := ServerConnectionHandshakePacket{
		PlayerId: playerId,
		BufferedChanges:buffer,
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(networkPacket); err != nil {
		fmt.Printf("Encoding Failed")
		return
	}

	//self.outboundTCPChannel <- buf.Bytes()
}

func (self *NetworkConnectionComponent) ConnectToDataChannel(data []byte) {

	//fmt.Println("Configuring ICE server")
	//config := webrtc.Configuration {
	//	ICEServers: []webrtc.ICEServer {
	//		{
	//			URLs: []string{"stun:stun.l.google.com:19302"},
	//		},
	//	},
	//}
	//
	//fmt.Println("Created Peer Connection")
	//peerConnection, err := webrtc.NewPeerConnection(config)
	//if err != nil {
	//	panic(err)
	//}
	//
	//// Set the handler for ICE connection state
	//// This will notify you when the peer has connected/disconnected
	//peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
	//	fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	//})
	//
	//peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
	//	fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())
	//
	//	// Register channel opening handling
	//	d.OnOpen(func() {
	//		fmt.Printf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds\n", d.Label(), d.ID())
	//	})
	//
	//	// Register text message handling
	//	d.OnMessage(func(msg webrtc.DataChannelMessage) {
	//		fmt.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
	//	})
	//})
	//
	//var handshake map[string]string
	//err = json.Unmarshal(data, handshake)
	//
	//if err != nil {
	//	// close connection
	//	fmt.Println("Handshake failed!")
	//	panic(err)
	//}
	//
	//offer := webrtc.SessionDescription{}
	//Decode(handshake["candidate"], &offer)
	//
	//err = peerConnection.SetRemoteDescription(offer)
	//if err != nil {
	//	panic(err)
	//}
	//
	//answer, err := peerConnection.CreateAnswer(nil)
	//if err != nil {
	//	panic(err)
	//}
	//
	//err = peerConnection.SetLocalDescription(answer)
	//if err != nil {
	//	panic(err)
	//}
	//
	//// send to client the answer
	//
	//handshake["candidate"] = Encode(answer)
	//marshal, err := json.Marshal(handshake)
	//
	//if err != nil {
	//	panic(err)
	//}
	//
	//self.outboundTCPChannel <- marshal
	//
	//self.PeerConnection = peerConnection
}


// Decode decodes the input from base64
// It can optionally unzip the input after decoding
func Decode(in string, obj interface{}) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		panic(err)
	}
}

func Encode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(b)
}
