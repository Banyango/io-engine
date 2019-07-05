package web

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/Banyango/io-engine/src/client"
	. "github.com/Banyango/io-engine/src/ecs"
	"github.com/Banyango/io-engine/src/server"
	"github.com/Banyango/socker"
	"net/url"
	"path"
	"sync"
	"syscall/js"
	"time"
)

type ConnectionStateType int

type NetworkedClientSystem struct {
	entities Storage

	ConnHandler *socker.SockerClient

	Client client.Client

	done chan struct{}

	ws       js.Value

	WebRTCConnection js.Value

	IsConnected      bool
	WaitingToResync  bool
	mux              sync.Mutex
	WorldStatePacket []*server.WorldState

	NetworkInstance Storage
}

func (self *NetworkedClientSystem) Init(w *World) {

	self.ConnHandler = socker.NewClient()

	self.NetworkInstance = NewStorage()

	self.Client = client.Client{}

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
		return false
	})

	go func() {
		// todo this needs to match the server address. Maybe build a service?
		u, err := url.Parse("ws://localhost:8081")

		if err != nil {
			fmt.Println("Error parsing URL")
			return
		}

		u.Path = path.Join(u.Path, "connect")

		js.Global().Get("console").Call("log", "Connecting to "+u.String())

		self.ws = js.Global().Get("WebSocket").New(u.String())
		self.ws.Set("binaryType", "arraybuffer")

		onopen := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			js.Global().Get("console").Call("log", "Creating WebRTC connection...")

			webrtcConnectionJs := js.Global().Get("window").Get("WebRTCConnection")

			if webrtcConnectionJs == js.Undefined() {
				log("Please include main.js in html page.")
			}

			self.WebRTCConnection = webrtcConnectionJs.New(self.ws)

			log("Adding onmessage..")

			self.WebRTCConnection.Get("sendChannel").Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				bytesRec := self.receiveWorldStateTick(args)
				w.BytesRec = bytesRec
				return nil
			}));

			self.WebRTCConnection.Get("sendChannel").Set("onclose", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				log("sendChannel closed")
				self.IsConnected = false
				return nil
			}))

			self.WebRTCConnection.Get("sendChannel").Set("onopen", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				self.onWebRTCConnectionOpened()
				return nil
			}))

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

/**
Networking Handlers
 */

func (self *NetworkedClientSystem) onWebRTCConnectionOpened() {
	log("sendChannel opened")
	self.IsConnected = true
	{
		// this is a hack and should be changed. advances the server state once the connection is opened.
		enc := js.Global().Get("TextEncoder").New()
		result := enc.Call("encode", "{\"event\":\"connectionOpened\"}")
		self.ws.Call("send", result)
	}
	enc := js.Global().Get("TextEncoder").New()
	result := enc.Call("encode", "{\"event\":\"spawn\"}")
	self.ws.Call("send", result)
}

func (self *NetworkedClientSystem) receiveWorldStateTick(args []js.Value) int {
	dataJSArray := js.Global().Get("Uint8Array").New(args[0].Get("data"))
	message := make([]byte, args[0].Get("data").Get("byteLength").Int())
	jsBuf := js.TypedArrayOf(message)
	jsBuf.Call("set", dataJSArray, 0)

	bytesRec := len(message)

	var packet server.ClientWorldStatePacket
	if err := gob.NewDecoder(bytes.NewReader(message)).Decode(&packet); err != nil {
		fmt.Println("Error in ClientWorldStatePacket", err)
	}

	var data server.WorldState
	if err := gob.NewDecoder(bytes.NewReader(packet.State)).Decode(&data); err != nil {
		fmt.Println("Error in WorldStatePacket", err)
	}

	if packet.RTT != nil {
		self.Client.HandleRTT(packet.RTT)
	}

	self.WorldStatePacket = append(self.WorldStatePacket, &data)
	jsBuf.Release()

	return bytesRec
}

func (self *NetworkedClientSystem) sendInputForCurrentFrame(world *World) {

	var buf bytes.Buffer

	data := server.ClientPacket{
		Tick:world.CurrentTick,
		Time:time.Now().UnixNano(),
		Input:world.Input.Player[0].ToBytes()[0],
	}

	err := gob.NewEncoder(&buf).Encode(data)

	if err != nil {
		fmt.Println(err)
	}

	jsBuf := js.TypedArrayOf(buf.Bytes())
	self.WebRTCConnection.Get("sendChannel").Call("send", jsBuf)
	jsBuf.Release()
}

func (self *NetworkedClientSystem) UpdateSystem(delta float64, world *World) {
	if world.IsResimulating {
		return
	}

	if self.IsDataChannelConnected() {

		world.Ping = self.Client.Ping;

		if self.WorldStatePacket != nil && len(self.WorldStatePacket) > 0 {
			for _, val := range self.WorldStatePacket {
				self.Client.HandleWorldStatePacket(val, world, &self.NetworkInstance)
			}
			self.WorldStatePacket = self.WorldStatePacket[:0]
		}

		if world.Input != nil && world.Input.Player[0] != nil {
			self.sendInputForCurrentFrame(world)
		} else {
			log("input nil")
		}

	}
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

func log(str ...interface{}) {
	js.Global().Get("console").Call("log", str...)
}

func (self *NetworkedClientSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{NetworkInstanceComponentType}
}

func (self *NetworkedClientSystem) AddToStorage(entity *Entity) {
	storages := map[int]*Storage{
		int(NetworkInstanceComponentType): &self.NetworkInstance,
	}
	AddComponentsToStorage(entity, storages)
}

func (self *NetworkedClientSystem) RemoveFromStorage(entity *Entity) {
	storages := map[int]*Storage{
		int(NetworkInstanceComponentType): &self.NetworkInstance,
	}
	RemoveComponentsFromStorage(entity, storages)
}

func (self *NetworkedClientSystem) IsDataChannelConnected() bool {
	return self.IsConnected
}

