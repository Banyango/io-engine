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

        this.sendChannel = this.pc.createDataChannel('foo', {
            // ordered:false,
            // maxRetransmits:1,
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

    setAnswer(sessionDescription) {
        this.pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(sessionDescription))))
    }
}

window.WebRTCConnection = WebRTCConnection;
