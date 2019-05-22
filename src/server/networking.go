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
	. "io-engine-backend/src/ecs"
	"sync"
	"time"
)

/*
----------------------------------------------------------------------------------------------------------------
Network Server Global
----------------------------------------------------------------------------------------------------------------
*/

type ServerGlobal struct {
	BufferedEntityChanges []EntityChange
	CurrentEntityChanges  []EntityChange
	CurrentState          WorldStatePacket
	mux                   sync.Mutex
	PlayerIndex           uint16
}

func (self *ServerGlobal) Id() int {
	return ServerGlobalType
}

func (self *ServerGlobal) ServerSpawn(ownerId uint16, world *World, prefabId byte, networkPrefab byte, buffer bool) {

	entity, err := world.PrefabData.CreatePrefab(int(prefabId))

	if err != nil {
		fmt.Println("Error Spawning entity: ", int(prefabId))
		return
	}

	entity.Id = world.FetchAndIncrementId()

	networkInstanceComponent := new(NetworkInstanceComponent)

	networkInstanceComponent.Data = NetworkData{}
	networkInstanceComponent.Data.OwnerId = ownerId
	networkInstanceComponent.Data.NetworkId = uint16(entity.Id)

	entity.Components[int(NetworkInstanceComponentType)] = networkInstanceComponent

	world.AddEntityToWorld(entity)

	self.NetworkSpawn(ownerId, uint16(entity.Id), prefabId, networkPrefab, buffer)

}

func (self *ServerGlobal) NetworkSpawn(ownerId uint16, networkId uint16, prefabId byte, networkPrefab byte, buffer bool) {
	self.mux.Lock()
	self.CreateChange(ownerId, networkId, prefabId, networkPrefab, buffer, InstantiatedEntityChangeType)
	self.mux.Unlock()
}

func (self *ServerGlobal) NetworkDestroy(ownerId uint16, networkId uint16, prefabId byte, networkPrefab byte, buffer bool) {
	self.mux.Lock()
	self.CreateChange(ownerId, networkId, prefabId, networkPrefab, buffer, DestroyedEntityChangeType)
	self.mux.Unlock()
}

func (self *ServerGlobal) CreateChange(ownerId uint16, networkId uint16, prefabId byte, networkPrefab byte, buffer bool, changeType EntityChangeType) EntityChange {
	change := EntityChange{OwnerId:ownerId, NetworkId: networkId, PrefabId: prefabId, NetworkPrefabId: networkPrefab, Type: changeType, Buffer: buffer}

	self.CurrentEntityChanges = append(self.CurrentEntityChanges, change)

	if buffer {
		self.BufferedEntityChanges = append(self.BufferedEntityChanges, change)
	}

	return change
}

func (self *ServerGlobal) NetworkSendUDP(data NetworkData) {
	self.mux.Lock()
	self.CurrentState.Updates = append(self.CurrentState.Updates, data)
	self.mux.Unlock()
}

func (self *ServerGlobal) CreateGlobal(world *World) {

}

func (self *ServerGlobal) Clear() {
	self.CurrentState.Updates = self.CurrentState.Updates[:0]
	self.CurrentEntityChanges = self.CurrentEntityChanges[:0]
}

func (self *ServerGlobal) FetchAndIncrementPlayerId() uint16 {
	self.mux.Lock()

	temp := self.PlayerIndex

	self.PlayerIndex++

	self.mux.Unlock()
	return temp

}

/*
----------------------------------------------------------------------------------------------------------------
Network Server System
----------------------------------------------------------------------------------------------------------------
*/

type ConnectionHandlerSystem struct {
	Connections          Storage
	NetworkInput	     Storage
}

func (self *ConnectionHandlerSystem) Init() {
	self.Connections = NewStorage()
	self.NetworkInput = NewStorage()

	dynamic.Register("ServerGlobal", func() interface{} {
		return &ServerGlobal{}
	})
}

func (self *ConnectionHandlerSystem) AddToStorage(entity *Entity) {
	for k := range entity.Components {
		component := entity.Components[k].(Component)

		if component.Id() == int(NetworkConnectionComponentType) {
			self.Connections.Components[entity.Id] = &component
		} else if component.Id() == int(NetworkInputComponentType) {
			self.NetworkInput.Components[entity.Id] = &component
		}
	}
}

func (*ConnectionHandlerSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{NetworkConnectionComponentType, NetworkInputComponentType}
}

func (self *ConnectionHandlerSystem) UpdateSystem(delta float64, world *World) {

	global := world.Globals[ServerGlobalType].(*ServerGlobal)

	var wsBytesToWrite []byte

	if len(global.CurrentEntityChanges) > 0 {

		// build the network packet.
		networkPacket := BufferedEntityPacket{
			Time:    time.Now().Unix(),
			Updates: global.CurrentEntityChanges,
		}

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)

		if err := enc.Encode(networkPacket); err != nil {
			fmt.Printf("Encoding Failed")
			return
		}

		wsBytesToWrite = buf.Bytes()

	}

	var udpBytesToWrite []byte
	if len(global.CurrentState.Updates) > 0 {
		var gameStateBuffer bytes.Buffer
		enc := gob.NewEncoder(&gameStateBuffer)
		if err := enc.Encode(global.CurrentState); err != nil {
			fmt.Printf("Encoding Failed")
			return
		}
		udpBytesToWrite = gameStateBuffer.Bytes()
	}

	for entity, _ := range self.Connections.Components {

		input := (*self.NetworkInput.Components[entity]).(*NetworkInputComponent)
		net := (*self.Connections.Components[entity]).(*NetworkConnectionComponent)

		net.WSConnHandler.Handle()

		// only send data changes if the webrtc connection is open.
		if net.IsDataChannelOpen {
			if len(wsBytesToWrite) > 0 {
				net.WSConnHandler.Write(wsBytesToWrite)
			}

			// handle webrtc messages
			select {
			case message, ok := <-net.UdpIn:
				if ok {
					//fmt.Println("Handling input")
					input.HandleClientInput(message)
				}
			default:
			}

			if net.DataChannel != nil && len(udpBytesToWrite) > 0 {
				err := net.DataChannel.Send(udpBytesToWrite)
				if err != nil {
					fmt.Println("Error Writing to data channel player:", net.PlayerId)
				}
			}
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

type EntityChange struct {
	OwnerId         uint16
	NetworkId       uint16
	NetworkPrefabId byte
	PrefabId        byte
	Type            EntityChangeType
	Buffer          bool
}

// Serialized Bytes of the entity and components.
type NetworkData struct {
	OwnerId   uint16
	NetworkId uint16
	Data map[int][]byte
}

type WorldStatePacket struct {
	Updates []NetworkData
}

type ReadSyncUDP interface {
	ReadUDP(networkPacket *NetworkData)
}

type WriteSyncUDP interface {
	WriteUDP(networkPacket *NetworkData)
}

type NetworkInfo struct {
	PlayerId uint16
	NetworkInput
}

type NetworkInput struct {
	Up    bool
	Down  bool
	Right bool
	Left  bool
	X     bool
	C     bool
	//Mouse1 bool
	//Mouse2 bool
	//mousePos
}

func (self *NetworkInput) ToBytes() []byte {
	value := byte(0)

	if self.Up {
		setBit(&value, 0)
	}

	if self.Down {
		setBit(&value, 1)
	}

	if self.Right {
		setBit(&value, 2)
	}

	if self.Left {
		setBit(&value, 3)
	}

	if self.X {
		setBit(&value, 4)
	}

	if self.C {
		setBit(&value, 5)
	}

	return []byte{value}
}

func NetworkInputFromBytes(value byte) NetworkInput {
	return NetworkInput{
		Up:    hasBit(value, 0),
		Down:  hasBit(value, 1),
		Right: hasBit(value, 2),
		Left:  hasBit(value, 3),
		X:     hasBit(value, 4),
		C:     hasBit(value, 5),
	}
}

func hasBit(n byte, pos uint) bool {
	val := n & (1 << pos)
	return val > 0
}

func setBit(n *byte, pos uint) {
	*n |= 1 << pos
}


type EntityChangeType byte

const (
	InstantiatedEntityChangeType EntityChangeType = iota
	DestroyedEntityChangeType
	StateUpdatedEntityChangeType
)

/*
----------------------------------------------------------------------------------------------------------------
	Network Connection Component - Handles the Websocket and WebRTC connections.
----------------------------------------------------------------------------------------------------------------
*/

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

func (self *NetworkConnectionComponent) Clone() Component {
	return new(NetworkConnectionComponent)
}

func (self *NetworkConnectionComponent) ConnectToDataChannel(data []byte) {

	self.UdpIn = make(chan []byte)

	fmt.Println("Configuring ICE server")
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
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

		self.DataChannel = d

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
		fmt.Println("Adding candidate: ", handshake["candidate"])
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

func (self *NetworkConnectionComponent) SendBufferedEntites(buffer []EntityChange) {

	networkPacket := ServerConnectionHandshakePacket{
		PlayerId:        self.PlayerId,
		BufferedChanges: buffer,
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
		fmt.Println("Decode Failure")
		panic(err)
	}
}

func Encode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		fmt.Println("Encode Failure")
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(b)
}

/*
----------------------------------------------------------------------------------------------------------------
Network Instance Component
----------------------------------------------------------------------------------------------------------------
*/

type NetworkInstanceComponent struct {
	Data NetworkData
}

func (*NetworkInstanceComponent) Id() int {
	return int(NetworkInstanceComponentType)
}

func (self *NetworkInstanceComponent) CreateComponent() {
	self.Data.Data = make(map[int][]byte)
}

func (*NetworkInstanceComponent) DestroyComponent() {

}

func (*NetworkInstanceComponent) Clone() Component {
	return new(NetworkInstanceComponent)
}

type NetworkInstanceDataCollectionSystem struct {
	NetworkInstances Storage
}

func (self *NetworkInstanceDataCollectionSystem) Init() {
	self.NetworkInstances = NewStorage()
}

func (self *NetworkInstanceDataCollectionSystem) AddToStorage(entity *Entity) {

	keys := map[int]*Storage {
		int(NetworkInstanceComponentType):&self.NetworkInstances,
	}

	AddComponentsToStorage(entity, keys)
}

func (self *NetworkInstanceDataCollectionSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{NetworkInstanceComponentType}
}

func (self *NetworkInstanceDataCollectionSystem) UpdateSystem(delta float64, world *World) {

	global := world.Globals[ServerGlobalType].(*ServerGlobal)

	for entity, _ := range self.NetworkInstances.Components {

		instance := (*self.NetworkInstances.Components[entity]).(*NetworkInstanceComponent)

		entity := world.Entities[entity]

		for _,val:= range entity.Components {
			if syncVar, ok := val.(WriteSyncUDP); ok {
				syncVar.WriteUDP(&instance.Data)
			}
		}

		global.NetworkSendUDP(instance.Data)

	}
}

// todo there is an issue with creating client side entities whereby the id will collide with the server id of other network entities.



