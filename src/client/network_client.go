package client

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/Banyango/socker"
	"io-engine-backend/src/server"
	. "io-engine-backend/src/shared"
	"net/url"
	"path"
	"sync"
	"syscall/js"
)

type ConnectionStateType int

//const (
//	ConnectionState_NotConnected ConnectionStateType = iota
//	ConnectionState_WSConnecting
//	ConnectionState_CreatingDataChannel
//	ConnectionState_Connected
//)
//
//
//type PreWebSocketHandshake struct {
//
//}
//
//func (self *PreWebSocketHandshake) HandleWSMessage(message []byte, system *NetworkedClientSystem) error {
//	var packet server.ServerConnectionHandshakePacket
//
//	if err := gob.NewDecoder(bytes.NewReader(message)).Decode(&packet); err != nil {
//		log(fmt.Sprintln("Error in handshake packet!"))
//		return err
//	}
//
//	system.PlayerId = int(packet.PlayerId)
//
//	log(fmt.Sprintln("Creating webrtc channel"))
//	system.SetupWebRTC()
//
//	//create a fake packet with the buffered changes
//	system.Buffered = packet.BufferedChanges
//
//	log(fmt.Sprintln("Player Id", system.PlayerId))
//	log(fmt.Sprintln("Is ConnectionState", system.ConnectionState))
//
//	self.NextState(system)
//
//	return nil
//}
//
//func (*PreWebSocketHandshake) NextState(system *NetworkedClientSystem) {
//	system.ConnectionState = &WaitForDataChannelHandshake{}
//}

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
	PlayerId    int
	ConnHandler *socker.SockerClient

	done chan struct{}
	data chan server.BufferedEntityPacket
	addr string

	doc      js.Value
	canvasEl js.Value
	ctx      js.Value
	ws       js.Value
	im       js.Value

	WebRTCConnection js.Value

	Packets          []server.BufferedEntityPacket
	Buffered         []server.EntityChange
	mux              sync.Mutex
}


func (self *NetworkedClientSystem) Init() {

	self.ConnHandler = socker.NewClient()

	// handle init handshake
	self.ConnHandler.Add(func(message []byte) bool {

		log("-- Handling init")
		var packet server.ServerConnectionHandshakePacket

		if err := gob.NewDecoder(bytes.NewReader(message)).Decode(&packet); err != nil {
			log(fmt.Sprintln("Error in handshake packet!"))
		}

		self.PlayerId = int(packet.PlayerId)

		js.Global().Get("console").Call("log", "Creating WebRTC connection...")

		self.SetupWebRTC()

		//create a fake packet with the buffered changes
		self.Buffered = packet.BufferedChanges

		log(fmt.Sprintln("Player Id", self.PlayerId))

		self.SetupWebRTC()

		return true
	})

	// handle webrtc answer
	self.ConnHandler.Add(func(message []byte) bool {

		log("-- Handling answer")
		var handshake map[string]string
		err := json.Unmarshal(message, &handshake)

		if err != nil {
			log("handshake failure ")
			panic(err)
		}

		log("-- Setting answer")
		self.WebRTCConnection.Call("setAnswer", handshake["answer"])

		return true
	})

	// handle normal TCP packets.
	self.ConnHandler.Add(func(message []byte) bool {
		var packet server.BufferedEntityPacket

		if err := gob.NewDecoder(bytes.NewReader(message)).Decode(&packet); err != nil {
			fmt.Println("Error in Packet",err)
		}

		self.Packets = append(self.Packets, packet)

		return false
	})

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

			enc := js.Global().Get("TextEncoder").New()

			result := enc.Call("encode", "Give me Entities")

			self.ws.Call("send", result)

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

			err := self.ConnHandler.Handle(message)

			if err != nil {
				fmt.Println("Error occurred")
				jsBuf.Release()
			}

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
	log("Creating WebRTCConnection..")
	webrtcConnectionJs := js.Global().Get("window").Get("WebRTCConnection")

	if webrtcConnectionJs == js.Undefined() {
		log("Please include main.js in html page.")
	}

	self.WebRTCConnection = webrtcConnectionJs.New(self.ws)

	log("Adding onmessage..")
	self.WebRTCConnection.Get("sendChannel").Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		log("got a message!")
		return nil;
	}));
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

func Decode(in string, obj interface{}) {
	log("Decoding")
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		log("base64 error")
		panic(err)
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		log("unmarshal error")
		panic(err)
	}
}

func getElementByID(id string) js.Value {
	return js.Global().Get("document").Call("getElementById", id)
}

func handleError(err error) {
	js.Global().Get("console").Call("log",err.Error())
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



