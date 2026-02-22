// WebRTC client for vbrowser
(function() {
    'use strict';

    const video = document.getElementById('video');
    const status = document.getElementById('status');
    
    let ws = null;
    let pc = null;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 5;

    function updateStatus(message, className) {
        status.textContent = message;
        status.className = className;
    }

    function connect() {
        updateStatus('● Connecting...', 'connecting');

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        
        ws = new WebSocket(wsUrl);

        ws.onopen = () => {
            console.log('WebSocket connected');
            reconnectAttempts = 0;
            initWebRTC();
        };

        ws.onmessage = async (event) => {
            try {
                const msg = JSON.parse(event.data);
                await handleMessage(msg);
            } catch (err) {
                console.error('Failed to handle message:', err);
            }
        };

        ws.onerror = (err) => {
            console.error('WebSocket error:', err);
        };

        ws.onclose = () => {
            console.log('WebSocket closed');
            updateStatus('● Disconnected', 'disconnected');
            
            if (pc) {
                pc.close();
                pc = null;
            }

            // Attempt reconnect
            if (reconnectAttempts < maxReconnectAttempts) {
                reconnectAttempts++;
                const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
                console.log(`Reconnecting in ${delay}ms (attempt ${reconnectAttempts})`);
                setTimeout(connect, delay);
            } else {
                updateStatus('● Connection failed', 'disconnected');
            }
        };
    }

    async function initWebRTC() {
        try {
            // Create peer connection
            pc = new RTCPeerConnection({
                iceServers: [
                    { urls: 'stun:stun.l.google.com:19302' }
                ]
            });

            // Handle incoming tracks
            pc.ontrack = (event) => {
                console.log('Received track:', event.track.kind);
                if (event.track.kind === 'video') {
                    video.srcObject = event.streams[0];
                    updateStatus('● Connected', 'connected');
                }
            };

            // Handle ICE candidates
            pc.onicecandidate = (event) => {
                if (event.candidate) {
                    send({
                        type: 'candidate',
                        candidate: event.candidate.toJSON()
                    });
                }
            };

            // Handle connection state
            pc.onconnectionstatechange = () => {
                console.log('Connection state:', pc.connectionState);
                switch (pc.connectionState) {
                    case 'connected':
                        updateStatus('● Connected', 'connected');
                        break;
                    case 'disconnected':
                    case 'failed':
                        updateStatus('● Disconnected', 'disconnected');
                        break;
                    case 'connecting':
                        updateStatus('● Connecting...', 'connecting');
                        break;
                }
            };

            // Create and send offer
            const offer = await pc.createOffer();
            await pc.setLocalDescription(offer);

            send({
                type: 'offer',
                sdp: pc.localDescription
            });

        } catch (err) {
            console.error('WebRTC initialization failed:', err);
            updateStatus('● Connection failed', 'disconnected');
        }
    }

    async function handleMessage(msg) {
        console.log('Received message:', msg.type);

        switch (msg.type) {
            case 'answer':
                if (pc && msg.sdp) {
                    await pc.setRemoteDescription(new RTCSessionDescription(msg.sdp));
                }
                break;

            case 'candidate':
                if (pc && msg.candidate) {
                    await pc.addIceCandidate(new RTCIceCandidate(msg.candidate));
                }
                break;

            case 'error':
                console.error('Server error:', msg.error);
                updateStatus('● Error: ' + msg.error, 'disconnected');
                break;

            default:
                console.warn('Unknown message type:', msg.type);
        }
    }

    function send(msg) {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify(msg));
        } else {
            console.error('WebSocket not ready');
        }
    }

    // Start connection
    connect();

})();
