package server

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	. "github.com/Banyango/io-engine/src/ecs"
	"github.com/Banyango/socker"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"
	"github.com/thoas/go-funk"
	"net/http"
	"sync"
)

const (
	SEND_TICK_RATE = 0.05
)

type Server struct {
	CurrentState WorldStatePacket
	Buffered     []*NetworkData
	mux          sync.Mutex
	PlayerIndex  uint16
	NetworkIndex uint16
	Clients      []*ClientConnection
	deltaCounter float64
	World        *World
}

func (self *Server) EntityWasSpawned(entity *Entity) {
	component := entity.Components[int(NetworkInstanceComponentType)].(*NetworkInstanceComponent)

	networkData := self.SerializeEntity(entity)
	networkData.NetworkId = component.NetworkId
	networkData.OwnerId = component.OwnerId
	networkData.PrefabId = uint16(component.PrefabId)

	fmt.Println("network spawn {owner=", component.OwnerId, " networkId=", component.NetworkId, " prefab=", component.PrefabId, "}")

	self.CurrentState.Created = append(self.CurrentState.Created, networkData)

	self.Buffered = append(self.Buffered, networkData)
}

func (self *Server) EntityWasDestroyed(entity int64) {
	if id, found := self.FindNetworkId(entity); found {
		indexToRemove := -1

		for index, val := range self.Buffered {
			if val.NetworkId == id {
				indexToRemove = index
				break
			}
		}

		if indexToRemove != -1 {
			self.Buffered = append(self.Buffered[:indexToRemove], self.Buffered[indexToRemove+1:]...)
		}

		self.CurrentState.Destroyed = append(self.CurrentState.Destroyed, int(id))
	}
}

func (self *Server) Ws(writer http.ResponseWriter, request *http.Request) {

	upgrader := websocket.Upgrader{}

	fmt.Println("Client is connecting..")

	conn, err := upgrader.Upgrade(writer, request, nil)

	if err != nil {
		fmt.Println("Error Occurred")
		return
	}

	self.createClientConnection(conn)

}

func (self *Server) SerializeEntity(entity *Entity) *NetworkData {

	data := NetworkData{Data: map[int][]byte{}}

	for i := range entity.Components {
		if val, ok := entity.Components[i].(WriteSyncUDP); ok {
			val.WriteUDP(&data)
		}
	}

	return &data
}

func (self *NetworkData) UpdateEntity(entity *Entity) {
	if entity.Components != nil {
		for _, comp := range entity.Components {
			if val, ok := comp.(ReadSyncUDP); ok {
				val.ReadUDP(self)
			}
		}
	}
}

func (self *NetworkData) DeserializeNewEntity(world *World, isPeer bool) *Entity {

	world.Log.LogInfo("deserializing entity")

	prefabId := int(self.PrefabId)

	if isPeer {
		prefabId = prefabId + 1
	}

	entity, err := world.PrefabData.CreatePrefab(prefabId)

	if err != nil {
		world.Log.LogInfo("error creating prefab")
		return nil
	}

	if entity.Components != nil {
		for _, comp := range entity.Components {
			if val, ok := comp.(ReadSyncUDP); ok {
				world.Log.LogJson("Reading update for", val)
				val.ReadUDP(self)
			}
		}
	}

	return &entity
}

func (self *Server) createClientConnection(conn *websocket.Conn) {

	clientConn := new(ClientConnection)

	self.AddClient(clientConn)

	clientConn.WSConnHandler = socker.NewClientConnection(conn)
	clientConn.WSConnHandler.Connection.SetCloseHandler(func(code int, text string) error {
		clientConn.Close(self.World)
		self.RemoveClient(clientConn)
		return nil
	})
	self.World.Mux.Lock()

	clientConn.PlayerId = self.FetchAndIncrementPlayerId()
	self.World.Input.Player[PlayerId(clientConn.PlayerId)] = NewInput()
	self.World.Mux.Unlock()

	fmt.Println("Client Given playerId: ", clientConn.PlayerId)

	clientConn.WSConnHandler.Add(func(message []byte) bool {
		fmt.Println("Sending Buffered Entities to -> player ", clientConn.PlayerId)
		clientConn.SendHandshakePacket(self.World.CurrentTick, self.Buffered)
		return true
	})

	clientConn.WSConnHandler.Add(func(message []byte) bool {
		if clientConn.PeerConnection == nil {
			fmt.Println("Setting up webrtc data channel -> player ", clientConn.PlayerId)
			clientConn.ConnectToDataChannel(message, func() {
				clientConn.Close(self.World)
				self.RemoveClient(clientConn)
			})
		}

		fmt.Println("Handling messages -> player", clientConn.PlayerId, string(message))
		clientConn.HandleSignal(message)

		return clientConn.IsDataChannelOpen
	})

	clientConn.WSConnHandler.Add(func(message []byte) bool {
		var handshake map[string]interface{}
		err := json.Unmarshal(message, &handshake)

		if err != nil {
			// close connection
			// todo handle this better. Unsubscribe the player.
			fmt.Println("messaging failed!")
			panic(err)
		}

		if val, ok := handshake["event"]; ok {
			if val == "spawn" {
				entity, err := self.World.PrefabData.CreatePrefab(0)

				if err != nil {
					// close connection
					// todo handle this better. Unsubscribe the player.
					fmt.Println("Handshake failed!")
					panic(err)
				}

				fmt.Println("spawn -> player ", clientConn.PlayerId)
				self.mux.Lock()
				defer self.mux.Unlock()
				networkInstanceComponent := new(NetworkInstanceComponent)
				networkInstanceComponent.OwnerId = clientConn.PlayerId
				networkInstanceComponent.NetworkId = self.FetchAndIncrementNetworkId()
				networkInstanceComponent.PrefabId = entity.PrefabId
				entity.Components[int(NetworkInstanceComponentType)] = networkInstanceComponent

				self.World.Mux.Lock()
				defer self.World.Mux.Unlock()
				self.World.Spawn(entity)

			} else if val == "resync" {
				fmt.Println("resync -> player ", clientConn.PlayerId)
				clientConn.SendHandshakePacket(self.World.CurrentTick, self.Buffered)
			}
		}

		return false
	})

	go clientConn.WSConnHandler.ReadPump()
	go clientConn.WSConnHandler.WritePump()

}

func (self *Server) Clear() {
	self.CurrentState.Updates = self.CurrentState.Updates[:0]
	self.CurrentState.Created = self.CurrentState.Created[:0]
	self.CurrentState.Destroyed = self.CurrentState.Destroyed[:0]
}

func (self *Server) ClearUpdates() {
	self.CurrentState.Updates = self.CurrentState.Updates[:0]
}

func (self *Server) FetchAndIncrementPlayerId() PlayerId {

	temp := self.PlayerIndex

	self.PlayerIndex++

	return PlayerId(temp)
}

func (self *Server) FetchAndIncrementNetworkId() uint16 {

	temp := self.NetworkIndex

	self.NetworkIndex++

	return temp
}

func (self *Server) HandleIncomingData(delta float64) {
	for _, client := range self.Clients {

		// Handle WS messages
		client.WSConnHandler.Handle()

		if client.IsDataChannelOpen {
			// Handle WebRTC messages
			select {
			case message, ok := <-client.UdpIn:
				if ok {
					if len(message) > 1 {
						tick := int64(binary.LittleEndian.Uint64(message))

						if tick < self.World.CurrentFrameTime {
							self.World.ResetInput(client.PlayerId)
						}

						self.World.SetFutureInput(tick, message[8], client.PlayerId)
					}
				}
			default:
			}
		} else {
			self.World.ResetInput(client.PlayerId)
		}
	}
}

func (self *Server) SendNetworkData(delta float64) {

	self.deltaCounter += delta

	if self.deltaCounter < SEND_TICK_RATE {
		self.ClearUpdates()
		return
	}

	var udpBytesToWrite []byte
	var gameStateBuffer bytes.Buffer
	enc := gob.NewEncoder(&gameStateBuffer)

	self.CurrentState.Tick = self.World.CurrentTick

	if err := enc.Encode(self.CurrentState); err != nil {
		fmt.Printf("Encoding Failed")
		return
	}
	udpBytesToWrite = gameStateBuffer.Bytes()

	for _, client := range self.Clients {
		// only send data changes if the webrtc connection is open.
		if client.IsDataChannelOpen {
			if client.DataChannel != nil && len(udpBytesToWrite) > 0 {
				err := client.DataChannel.Send(udpBytesToWrite)
				if err != nil {
					fmt.Println("Error Writing to data channel player:", client.PlayerId)
				}
			}
		}

	}

	if self.deltaCounter > SEND_TICK_RATE {
		self.deltaCounter = 0
	}

	// todo right now send state but this needs to be rm when deltas are brought in.
	self.Clear()
}

func (self *Server) AddClient(connection *ClientConnection) {
	self.mux.Lock()
	self.Clients = append(self.Clients, connection)
	self.mux.Unlock()
}

func (self *Server) RemoveClient(connection *ClientConnection) {
	self.mux.Lock()
	indexOf := funk.IndexOf(self.Clients, connection)
	self.Clients = append(self.Clients[:indexOf], self.Clients[indexOf+1:]...)
	self.mux.Unlock()
}

func (self *Server) FindNetworkId(entityId int64) (uint16, bool) {
	if ent, ok := self.World.Entities[entityId]; ok {
		if c, ok := ent.Components[int(NetworkInstanceComponentType)]; ok {
			if netInstance, ok := c.(*NetworkInstanceComponent); ok {
				return netInstance.NetworkId, true
			}
		}
	}
	return 0, false
}

type ClientConnection struct {
	PlayerId PlayerId
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
	Resync         bool

	SimulateFaster bool
	Data           NetworkData

}

func (self *ClientConnection) ConnectToDataChannel(data []byte, onClose func()) {

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
		if connectionState == webrtc.ICEConnectionStateDisconnected {
			self.IsDataChannelOpen = false
		} else if connectionState == webrtc.ICEConnectionStateConnected {
			self.IsDataChannelOpen = true
		}
	})

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

		self.DataChannel = d

		d.OnOpen(func() {
			fmt.Printf("Opened DataChannel player{%d} %d\n", self.PlayerId, d.ID())
			self.IsDataChannelOpen = true
		})

		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			self.UdpIn <- msg.Data
		})

		d.OnClose(func() {
			fmt.Printf("Closing connection ->", self.PlayerId)
		})
	})

	self.PeerConnection = peerConnection
}

func (self *ClientConnection) HandleSignal(data []byte) {

	var handshake map[string]interface{}
	err := json.Unmarshal(data, &handshake)

	if err != nil {
		// close connection
		// todo handle this better. Unsubscribe the player.
		fmt.Println("Handshake failed!")
		panic(err)
	}

	if val, ok := handshake["offer"]; ok {
		fmt.Println("Rec Offer: ", handshake["offer"])
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

func (self *ClientConnection) SendHandshakePacket(tick int64, buffered []*NetworkData) {

	networkPacket := ServerConnectionHandshakePacket{
		PlayerId:   self.PlayerId,
		ServerTick: tick,
		State:      buffered,
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(networkPacket); err != nil {
		fmt.Printf("Encoding Failed")
		return
	}

	self.WSConnHandler.Write(buf.Bytes())

}

func (self *ClientConnection) Close(world *World) {
	world.Mux.Lock()
	defer world.Mux.Unlock()

	for id, ent := range world.Entities {
		if comp, ok := ent.Components[int(NetworkInstanceComponentType)]; ok {
			if net, ok := comp.(*NetworkInstanceComponent); ok {
				if net.OwnerId == self.PlayerId {
					world.ToDestroy = append(world.ToDestroy, id)
				}
			}
		}
	}
	fmt.Println("Removing player ", self.PlayerId)
	delete(world.Input.Player, self.PlayerId)

	err := self.PeerConnection.Close()

	if err != nil {
		fmt.Println(err)
	}

	err = self.DataChannel.Close()

	if err != nil {
		fmt.Println(err)
	}
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
