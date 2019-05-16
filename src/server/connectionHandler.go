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
	"time"
)


/*
----------------------------------------------------------------------------------------------------------------
Network Server Global
----------------------------------------------------------------------------------------------------------------
*/

type ServerGlobal struct {
	BufferedChanges []EntityChange
	CurrentChanges  []EntityChange
	mux             sync.Mutex
}

type EntityChange struct {
	NetworkId       uint16
	NetworkPrefabId byte
	PrefabId        byte
	Type            EntityChangeType
	Buffer          bool
}

type WorldStatePacket struct {
	NetworkId uint16
	PositionX uint32
	PositionY uint32
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

	self.CurrentChanges = append(self.CurrentChanges, change)

	if buffer {
		self.BufferedChanges = append(self.BufferedChanges, change)
	}

	return change
}

func (self *ServerGlobal) CreateGlobal(world *World) {

}

func (self *ServerGlobal) Clear() {
	self.CurrentChanges = self.CurrentChanges[:0]
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

	global := world.Globals[ServerGlobalType].(*ServerGlobal)

	var bytesToWrite []byte

	if len(global.CurrentChanges) > 0 {

		// build the network packet.
		networkPacket := BufferedEntityPacket{
			Time:    time.Now().Unix(),
			Updates: global.CurrentChanges,
		}

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)

		if err := enc.Encode(networkPacket); err != nil {
			fmt.Printf("Encoding Failed")
			return
		}

		bytesToWrite = buf.Bytes()

	}

	for entity, _ := range self.Connections.Components {
		net := (*self.Connections.Components[entity]).(*NetworkConnectionComponent)
		net.WSConnHandler.Handle()

		// only send data changes if the webrtc connection is open.
		if net.IsDataChannelOpen {
			if len(bytesToWrite) > 0 {
				net.WSConnHandler.Write(bytesToWrite)
			}

			// handle webrtc messages
			select {
			case message, ok := <-net.UdpIn:
				if ok {
					print(message)
				}
			default:
			}

			// send webrtc messages
			//net.DataChannel.Send()
		}

	}

	global.Clear()
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
	WSConnHandler *socker.SockerClientConnection

	// WebRTC channel
	UdpIn chan []byte

	// Udp - unreliable
	PeerConnection    *webrtc.PeerConnection
	DataChannel       *webrtc.DataChannel
	Offer             webrtc.SessionDescription
	AnswerSent        bool
	IsDataChannelOpen bool

}

func (*NetworkConnectionComponent) Id() int {
	return int(NetworkConnectionComponentType)
}

func (self *NetworkConnectionComponent) CreateComponent() {

}

func (*NetworkConnectionComponent) DestroyComponent() {

}

func (self *NetworkConnectionComponent) ConnectToDataChannel(data []byte) {

	self.UdpIn = make(chan []byte)

	fmt.Println("Configuring ICE server")
	config := webrtc.Configuration {
		ICEServers: []webrtc.ICEServer {
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	fmt.Println("Created Peer Connection")
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

		d.OnOpen(func() {
			self.IsDataChannelOpen = true
		})

		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			self.UdpIn <- msg.Data
		})
	})

	self.PeerConnection = peerConnection
}

func (self *NetworkConnectionComponent) HandleSignal(data []byte) {

	var handshake map[string]interface{}
	err := json.Unmarshal(data, &handshake)

	if err != nil {
		// close connection
		// todo handle this better. Unsubscribe the player.
		fmt.Println("Handshake failed!")
		panic(err)
	}

	if val, ok := handshake["offer"]; ok {
		offer := webrtc.SessionDescription{}
		Decode(val.(string), &offer)

		err = self.PeerConnection.SetRemoteDescription(offer)
		if err != nil {
			panic(err)
		}

		answer, err := self.PeerConnection.CreateAnswer(nil)

		if err != nil {
			panic(err)
		}

		err = self.PeerConnection.SetLocalDescription(answer)
		if err != nil {
			panic(err)
		}

		handshake["answer"] = Encode(answer)
		marshal, err := json.Marshal(handshake)

		if err != nil {
			panic(err)
		}

		self.WSConnHandler.Write(marshal)

	} else if val, ok := handshake["candidate"]; ok {
		fmt.Println("Adding candidate: ",handshake["candidate"])
		candidateInit := webrtc.ICECandidateInit{}

		err = json.Unmarshal([]byte(val.(string)), &candidateInit)
		if err != nil {
			panic(err)
		}

		err := self.PeerConnection.AddICECandidate(candidateInit)
		if err != nil {
			fmt.Println("Ice candidate error")
			panic(err)
		}
	}
}

func (self *NetworkConnectionComponent) GameMessage(message []byte) {
	// this is a TCP message from a running client.

}

func (self *NetworkConnectionComponent) SendBufferedEntites(playerId uint16, buffer []EntityChange) {

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

	self.WSConnHandler.Write(buf.Bytes())

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
