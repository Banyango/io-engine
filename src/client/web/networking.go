package web

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/Banyango/socker"
	. "io-engine-backend/src/ecs"
	"io-engine-backend/src/server"
	"net/url"
	"path"
	"sync"
	"syscall/js"
)

type ConnectionStateType int

type NetworkedClientSystem struct {

	entities Storage

	PlayerId    int
	ConnHandler *socker.SockerClient

	done chan struct{}
	data chan server.WorldStatePacket
	addr string

	doc      js.Value
	canvasEl js.Value
	ctx      js.Value
	ws       js.Value
	im       js.Value

	WebRTCConnection js.Value

	IsConnected      bool
	Packets          []server.WorldStatePacket
	State            server.WorldStatePacket
	mux              sync.Mutex
	WorldStatePacket []server.WorldStatePacket
}


func (self *NetworkedClientSystem) Init(w *World) {

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
		var packet server.ServerConnectionHandshakePacket

		if err := gob.NewDecoder(bytes.NewReader(message)).Decode(&packet); err != nil {
			fmt.Println("Error in Packet",err)
		}

		self.PlayerId = int(packet.PlayerId)

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

		dataJSArray := js.Global().Get("Uint8Array").New(args[0].Get("data"))
		message := make([]byte, args[0].Get("data").Get("byteLength").Int())

		jsBuf := js.TypedArrayOf(message)
		jsBuf.Call("set", dataJSArray, 0)

		var packet server.WorldStatePacket

		if err := gob.NewDecoder(bytes.NewReader(message)).Decode(&packet); err != nil {
			fmt.Println("Error in Packet",err)
		}

		self.WorldStatePacket = append(self.WorldStatePacket, packet)

		jsBuf.Release()

		return nil
	}));

	self.WebRTCConnection.Get("sendChannel").Set("onclose", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		log("sendChannel closed")
		self.IsConnected = false
		return nil
	}))

	self.WebRTCConnection.Get("sendChannel").Set("onopen", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		log("sendChannel opened")
		self.IsConnected = true
		return nil
	}))
}

// Encode encodes the input in base64
// It can optionally zip the input before encoding
func Encode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		log("Client Encode Failure")
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

func log(str... interface{}) {
	js.Global().Get("console").Call("log",str...)
}

func (self *NetworkedClientSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{}
}

func (self *NetworkedClientSystem) AddToStorage(entity *Entity) {

}

func (self *NetworkedClientSystem) RemoveFromStorage(entity *Entity) {

}

func (self *NetworkedClientSystem) UpdateSystem(delta float64, world *World) {

	if world.IsResimulating {
		return
	}

	if len(self.WorldStatePacket) > 0 {
		for _, val := range self.WorldStatePacket {

			for _, created := range val.Created {
				world.AddEntityToWorld(*created.DeserializeNewEntity(world))
			}

			for _, destroyed := range val.Destroyed {
				world.RemoveEntity(int64(destroyed))
			}

			resimulateRequired := false
			for _, update := range val.Updates {
				temp := update.DeserializeNewEntity(world)
				if !world.CompareEntitiesAtTick(val.Tick, temp) {
					resimulateRequired = true
					break
				}
			}

			if resimulateRequired {
				world.ResetToTick(val.Tick)

				for _, update := range val.Updates {
					update.UpdateEntity(world.Entities[int64(update.NetworkId)])
				}

				world.Resimulate(val.Tick)
			}
		}
	}

	if self.IsDataChannelConnected() {

		jsBuf := js.TypedArrayOf(world.Input.Player[0].ToBytes())

		//log("sending input" + jsBuf.String())
		self.WebRTCConnection.Get("sendChannel").Call("send", jsBuf)

		jsBuf.Release()
	}
}

func (self *NetworkedClientSystem) IsDataChannelConnected() bool {
	return self.IsConnected
}


func (self *NetworkedClientSystem) GetPrefabId(change server.EntityChange) int {
	prefabIdToCreate := int(change.PrefabId)
	if uint16(self.PlayerId) != change.OwnerId {
		prefabIdToCreate = int(change.NetworkPrefabId)
	}
	return prefabIdToCreate
}

type ClientNetworkGlobal struct {
	Packets []server.WorldStatePacket
}


func (self *ClientNetworkGlobal) CreateGlobal(world *World) {
	self.Packets = []server.WorldStatePacket{}
}


