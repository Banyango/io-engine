package server

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/Banyango/socker"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc"
	. "io-engine-backend/src/ecs"
	"net/http"
	"sync"
)

const (
	SEND_TICK_RATE = 0.05
)

type Server struct {
	CurrentState WorldStatePacket
	mux          sync.Mutex
	PlayerIndex  uint16
	Clients      []*ClientConnection
	deltaCounter float64
	World        *World
}

func (self *Server) EntityWasSpawned(entity *Entity) {
	self.CurrentState.Created = append(self.CurrentState.Created, self.SerializeEntity(entity))
}

func (self *Server) EntityWasDestroyed(entity int64) {
	self.CurrentState.Destroyed = append(self.CurrentState.Destroyed, int(entity))
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

func (self *Server) SerializeEntity(entity *Entity) NetworkData {

	data := NetworkData{}

	for i := range entity.Components {
		if val, ok := entity.Components[i].(WriteSyncUDP); ok {
			val.WriteUDP(&data)
		}
	}

	return data
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

func (self *NetworkData) DeserializeNewEntity(world *World) *Entity {

	entity, err := world.PrefabData.CreatePrefab(int(self.PrefabId))

	if err != nil {
		fmt.Println("error creating prefab")
		return nil
	}

	if entity.Components != nil {
		for _, comp := range entity.Components {
			if val, ok := comp.(ReadSyncUDP); ok {
				val.ReadUDP(self)
			}
		}
	}

	return &entity
}

func (self *Server) createClientConnection(conn *websocket.Conn) {

	entity, err := self.World.PrefabData.CreatePrefab(0)

	if err != nil {
		fmt.Println("Error Occurred Creating Player Prefab")
		return
	}

	clientConn := new(ClientConnection)

	self.AddClient(clientConn)

	clientConn.WSConnHandler = socker.NewClientConnection(conn)
	clientConn.PlayerId = self.FetchAndIncrementPlayerId()

	networkInstanceComponent := new(NetworkInstanceComponent)
	networkInstanceComponent.PlayerId = PlayerId(clientConn.PlayerId)
	entity.Components[int(NetworkInstanceComponentType)] = networkInstanceComponent

	self.World.Spawn(entity)

	self.World.Input.Player[PlayerId(clientConn.PlayerId)] = NewInput()

	fmt.Println("Client entityId: ", entity.Id, " Given playerId: ", clientConn.PlayerId)

	clientConn.WSConnHandler.Add(func(message []byte) bool {
		fmt.Println("Sending Buffered Entities to -> player ", clientConn.PlayerId)
		clientConn.SendHandshakePacket()
		return true
	})

	// setup webrtc connection send offer
	clientConn.WSConnHandler.Add(func(message []byte) bool {
		fmt.Println("Setting up webrtc data channel -> player ", clientConn.PlayerId)
		clientConn.ConnectToDataChannel(message)
		return true
	})

	clientConn.WSConnHandler.Add(func(message []byte) bool {
		fmt.Println("Handling messages -> player", clientConn.PlayerId)
		clientConn.HandleSignal(message)
		return clientConn.IsDataChannelOpen
	})

	clientConn.WSConnHandler.Add(func(message []byte) bool {
		//clientConn.GameMessage(message)
		return false
	})

	go clientConn.WSConnHandler.ReadPump()
	go clientConn.WSConnHandler.WritePump()

}

func (self *Server) Clear() {
	self.CurrentState.Updates = self.CurrentState.Updates[:0]
}

func (self *Server) FetchAndIncrementPlayerId() uint16 {

	temp := self.PlayerIndex

	self.PlayerIndex++

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
					// Handle input bytes from client
					self.World.Input.Player[PlayerId(client.PlayerId)].InputFromBytes(message[0])
				}
			default:
			}
		}
	}
}

func (self *Server) SendNetworkData(delta float64) {

	self.deltaCounter += delta

	if self.deltaCounter < SEND_TICK_RATE {
		return
	}

	var udpBytesToWrite []byte
	if len(self.CurrentState.Updates) > 0 {
		var gameStateBuffer bytes.Buffer
		enc := gob.NewEncoder(&gameStateBuffer)
		if err := enc.Encode(self.CurrentState); err != nil {
			fmt.Printf("Encoding Failed")
			return
		}
		udpBytesToWrite = gameStateBuffer.Bytes()
	}

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

type ClientConnection struct {
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

	Data NetworkData
}

func (self *ClientConnection) ConnectToDataChannel(data []byte) {

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

		d.OnClose(func() {
			self.Close()
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

func (self *ClientConnection) SendHandshakePacket() {

	networkPacket := ServerConnectionHandshakePacket{
		PlayerId: self.PlayerId,
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(networkPacket); err != nil {
		fmt.Printf("Encoding Failed")
		return
	}

	self.WSConnHandler.Write(buf.Bytes())

}

func (self *ClientConnection) Close() {

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
