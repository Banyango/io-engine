function startup() {
    connectButton = document.getElementById('connectButton');

    connectButton.addEventListener('click', connectPeers, false);
}

function connectPeers() {
    var webSocket = new WebSocket("ws://localhost:8081/connect");

    webSocket.onopen = function(event) {
        console.log("connected from js!");
    };

    webSocket.onmessage = function (message) {
        let pc = new RTCPeerConnection({
            iceServers: [
                {
                    urls: 'stun:stun.l.google.com:19302'
                }
            ]
        })


    }
}