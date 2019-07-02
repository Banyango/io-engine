class WebRTCConnection {
    constructor(websocket) {

        this.webSocket = websocket;

        this.isConnected = false;

        console.log("[[js]: starting init]");
        this.pc = new RTCPeerConnection({
            iceServers: [
                {
                    urls: 'stun:stun.l.google.com:19302'
                }
            ]
        });
        console.log("[[js]: peer connection init]");

        this.sendChannel = this.pc.createDataChannel('foo' + this.uuid4(), {
            ordered:false,
            protocol:"udp",
            maxRetransmits:5,
            priority:"high",
        });
        this.sendChannel.binaryType = 'arraybuffer';

        console.log("[[js]: data channel created]");

        this.pc.oniceconnectionstatechange = e => console.log(`[js]: icestate:${this.pc.iceConnectionState}`);

        this.pc.onnegotiationneeded = async () => {
            try {
                await this.pc.setLocalDescription(await this.pc.createOffer());
                // send the offer to the other peer
                const enc = new TextEncoder();
                this.webSocket.send(enc.encode(JSON.stringify({"offer": btoa(JSON.stringify(this.pc.localDescription))})));
                console.log("[[js]: Offer sent..]");
            } catch (err) {
                console.error(err);
            }
        };

        this.pc.onicecandidate = event => {
            if (event.candidate !== null && event.candidate.candidate !== "") {
                // send the candidate to the remote peer
                const enc = new TextEncoder();
                console.log(JSON.stringify(event.candidate.candidate));
                this.webSocket.send(enc.encode(JSON.stringify({"candidate": JSON.stringify(event.candidate.toJSON())})));
                console.log("[[js]: ICE candidate sent..]");
            }
        };
    }
    uuid4() {
        const ho = (n, p) => n.toString(16).padStart(p, 0); /// Return the hexadecimal text representation of number `n`, padded with zeroes to be of length `p`
        const view = new DataView(new ArrayBuffer(16)); /// Create a view backed by a 16-byte buffer
        crypto.getRandomValues(new Uint8Array(view.buffer)); /// Fill the buffer with random data
        view.setUint8(6, (view.getUint8(6) & 0xf) | 0x40); /// Patch the 6th byte to reflect a version 4 UUID
        view.setUint8(8, (view.getUint8(8) & 0x3f) | 0x80); /// Patch the 8th byte to reflect a variant 1 UUID (version 4 UUIDs are)
        return `${ho(view.getUint32(0), 8)}-${ho(view.getUint16(4), 4)}-${ho(view.getUint16(6), 4)}-${ho(view.getUint16(8), 4)}-${ho(view.getUint32(10), 8)}${ho(view.getUint16(14), 4)}`; /// Compile the canonical textual form from the array data
    }
    setAnswer(sessionDescription) {
        this.pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(sessionDescription))))
    }
}

window.WebRTCConnection = WebRTCConnection;
