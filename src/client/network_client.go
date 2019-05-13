package client

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/pion/webrtc"
	"io-engine-backend/src/server"
	. "io-engine-backend/src/shared"
	"net/url"
	"path"
	"sync"
	"syscall/js"
)

type ConnectionStateType int

const (
	ConnectionState_NotConnected ConnectionStateType = iota
	ConnectionState_WSConnecting
	ConnectionState_CreatingDataChannel
	ConnectionState_Connected
)

type ConnectionState interface {
	HandleWSMessage(message []byte, system *NetworkedClientSystem) error
	NextState(system *NetworkedClientSystem)
}

type PreWebSocketHandshake struct {

}

func (self *PreWebSocketHandshake) HandleWSMessage(message []byte, system *NetworkedClientSystem) error {
	var packet server.ServerConnectionHandshakePacket

	if err := gob.NewDecoder(bytes.NewReader(message)).Decode(&packet); err != nil {
		log(fmt.Sprintln("Error in handshake packet!"))
		return err
	}

	system.PlayerId = int(packet.PlayerId)

	log(fmt.Sprintln("Creating webrtc channel"))
	system.SetupWebRTC()

	//create a fake packet with the buffered changes
	system.Buffered = packet.BufferedChanges

	log(fmt.Sprintln("Player Id", system.PlayerId))
	log(fmt.Sprintln("Is ConnectionState", system.ConnectionState))

	self.NextState(system)

	return nil
}

func (*PreWebSocketHandshake) NextState(system *NetworkedClientSystem) {
	system.ConnectionState = &WaitForDataChannelHandshake{}
}

type WaitForDataChannelHandshake struct {

}

func (*WaitForDataChannelHandshake) HandleWSMessage(message []byte, system *NetworkedClientSystem) error {

	var packet map[string]string

	if err := gob.NewDecoder(bytes.NewReader(message)).Decode(&packet); err != nil {
		log(fmt.Sprintln("Error in handshake packet!"))
		return err
	}


	return nil
}

func (*WaitForDataChannelHandshake) NextState(system *NetworkedClientSystem){

}

type Connected struct {

}

func (*Connected) HandleWSMessage(message []byte, system *NetworkedClientSystem) error {
	return nil
}

func (*Connected) NextState(system *NetworkedClientSystem) {

}

type NetworkedClientSystem struct {
	PlayerId        int
	ConnectionState ConnectionState

	done chan struct{}
	data chan server.BufferedEntityPacket
	addr string

	doc      js.Value
	canvasEl js.Value
	ctx      js.Value
	ws       js.Value
	im       js.Value

	Packets       []server.BufferedEntityPacket
	Buffered      []server.EntityChange
	mux           sync.Mutex
	UDPConnection *webrtc.PeerConnection
}


func (self *NetworkedClientSystem) Init() {

	self.ConnectionState = &PreWebSocketHandshake{}

	go func() {
		u, err := url.Parse("ws://localhost:8081")

		if err != nil {
			fmt.Println("Error parsing URL")
			return
		}

		u.Path = path.Join(u.Path, "connect")

		js.Global().Get("console").Call("log", "Connecting to " + u.String())

		self.ws = js.Global().Get("WebSocket").New(u.String())
		self.ws.Set("binaryType", "arraybuffer")
		onopen := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			js.Global().Get("console").Call("log", "ConnectionState to server!")
			return nil
		})
		defer onopen.Release()
		onmessage := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// Create a buffer and copy it into a wasm referenced slice so we can use it in golang
			// make sure to release the typed array when done or there's a memory leak.
			dataJSArray := js.Global().Get("Uint8Array").New(args[0].Get("data"))
			message := make([]byte, args[0].Get("data").Get("byteLength").Int())

			jsBuf := js.TypedArrayOf(message)
			jsBuf.Call("set", dataJSArray, 0)

			log(fmt.Sprintln("got message", self.PlayerId))

			if !self.ConnectionState {

				var packet server.ServerConnectionHandshakePacket

				if err := gob.NewDecoder(bytes.NewReader(message)).Decode(&packet); err != nil {
					log(fmt.Sprintln("Error in handshake packet!"))

				}

				self.PlayerId = int(packet.PlayerId)
				self.ConnectionState = true

				log(fmt.Sprintln("Creating webrtc channel"))
				self.SetupWebRTC()

				//create a fake packet with the buffered changes
				self.Buffered = packet.BufferedChanges

				log(fmt.Sprintln("Player Id", self.PlayerId))
				log(fmt.Sprintln("Is ConnectionState", self.ConnectionState))

			} else {

				var packet server.BufferedEntityPacket

				if err := gob.NewDecoder(bytes.NewReader(message)).Decode(&packet); err != nil {
					fmt.Println("Error in Packet",err)
					goto cleanup
				}

				self.Packets = append(self.Packets, packet)

			}

			cleanup:
				jsBuf.Release()

			return nil
		})
		defer onmessage.Release()
		self.ws.Set("onopen", onopen)
		self.ws.Set("onmessage", onmessage)

		<-self.done
	}()
}

func (self *NetworkedClientSystem) SetupWebRTC () {

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		handleError(err)
	}

	ordered := false
	options := &webrtc.DataChannelInit{
		Ordered:&ordered,
	}
	// Create DataChannel.
	sendChannel, err := pc.CreateDataChannel("data " + string(self.PlayerId), options)
	if err != nil {
		handleError(err)
	}
	sendChannel.OnClose(func() {
		fmt.Println("sendChannel has closed")
	})
	sendChannel.OnOpen(func() {
		fmt.Println("sendChannel has opened")
	})
	sendChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		log(fmt.Sprintf("Message from DataChannel %s payload %s", sendChannel.Label(), string(msg.Data)))
	})

	// Create offer
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		handleError(err)
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		handleError(err)
	}

	// Add handlers for setting up the connection.
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log(fmt.Sprint(state))
	})
	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			encodedDescr := Encode(pc.LocalDescription())

			handshake := map[string]interface{} {
				"candidate":encodedDescr,
			}

			result, err := json.Marshal(handshake)

			if err != nil {
				log("Cant perform handshake")
				panic(err)
			}

			self.ws.Call("send", result)
		}
	})
}

// Encode encodes the input in base64
// It can optionally zip the input before encoding
func Encode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(b)
}



func getElementByID(id string) js.Value {
	return js.Global().Get("document").Call("getElementById", id)
}

func handleError(err error) {
	js.Global().Get("console").Call("log","Unexpected error. Check console.")
	panic(err)
}

func log(str string) {
	js.Global().Get("console").Call("log",str)
}

func (self *NetworkedClientSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{
		RawInputGlobalType,
		InputGlobalType,
	}
}

func (self *NetworkedClientSystem) AddToStorage(entity Entity) {

}

func (self *NetworkedClientSystem) UpdateSystem(delta float64, world *World) {
	// Grab the messages from the channel
	// If there is one process it. If not continue client side prediction.

	if len(self.Buffered) > 0 {
		for j := range self.Buffered {
			change := self.Buffered[j]
			if change.Type == server.InstantiatedEntityChangeType {

				prefabIdToCreate := self.GetPrefabId(change)

				js.Global().Get("console").Call("log", "Creating buffered entity: ", prefabIdToCreate)

				if entity, err := world.PrefabData.CreatePrefab(prefabIdToCreate); err != nil {
					fmt.Println(err)
				} else {
					world.AddEntityToWorld(entity)
				}
			}
		}
		self.Buffered = nil
	}

	if len(self.Packets) > 0 {
		message := self.Packets[len(self.Packets)-1]
		for j := range message.Updates {
			change := message.Updates[j]
			if change.Type == server.InstantiatedEntityChangeType {

				prefabIdToCreate := self.GetPrefabId(change)

				js.Global().Get("console").Call("log", "Creating entity: ", prefabIdToCreate)

				if entity, err := world.PrefabData.CreatePrefab(prefabIdToCreate); err != nil {
					fmt.Println(err)
				} else {
					world.AddEntityToWorld(entity)
				}
			}

			// validate our state at time whatever is equal to client side.

			// was there a difference

			// yes - resimulate the position up till now using our input buffer.

		}
		self.Packets = self.Packets[:0]
	}

}

func (self *NetworkedClientSystem) GetPrefabId(change server.EntityChange) int {
	prefabIdToCreate := int(change.PrefabId)
	if uint16(self.PlayerId) != change.NetworkId {
		prefabIdToCreate = int(change.NetworkPrefabId)
	}
	return prefabIdToCreate
}



