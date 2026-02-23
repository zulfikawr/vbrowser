// WebRTC client for vbrowser
(function() {
    'use strict';

    const video = document.getElementById('video');
    const status = document.getElementById('status');
    const overlay = document.getElementById('overlay');
    const cursorEl = document.getElementById('cursor');
    const selectRes = document.getElementById('select-res');
    const selectScaling = document.getElementById('select-scaling');
    const selectFps = document.getElementById('select-fps');
    const selectBitrate = document.getElementById('select-bitrate');
    const btnApply = document.getElementById('btn-apply');
    
    function startPlayback() {
        video.muted = true;
        video.play().then(() => {
            overlay.classList.remove('visible');
            console.log('Autoplay successful (muted)');
        }).catch(err => {
            console.log('Autoplay blocked even when muted, showing overlay');
            overlay.classList.add('visible');
        });
    }

    overlay.addEventListener('click', () => {
        startPlayback();
    });

    function getMousePos(e) {
        const rect = video.getBoundingClientRect();
        
        if (selectScaling.value === 'fill') {
            const x = (e.clientX - rect.left) * (video.videoWidth / rect.width);
            const y = (e.clientY - rect.top) * (video.videoHeight / rect.height);
            return { x: Math.round(x), y: Math.round(y) };
        }

        const videoRatio = video.videoWidth / video.videoHeight;
        const elementRatio = rect.width / rect.height;

        let contentWidth, contentHeight, offsetX, offsetY;

        if (elementRatio > videoRatio) {
            contentHeight = rect.height;
            contentWidth = rect.height * videoRatio;
            offsetX = (rect.width - contentWidth) / 2;
            offsetY = 0;
        } else {
            contentWidth = rect.width;
            contentHeight = rect.width / videoRatio;
            offsetX = 0;
            offsetY = (rect.height - contentHeight) / 2;
        }

        const x = (e.clientX - rect.left - offsetX) * (video.videoWidth / contentWidth);
        const y = (e.clientY - rect.top - offsetY) * (video.videoHeight / contentHeight);

        return { x: Math.round(x), y: Math.round(y) };
    }

    video.addEventListener('mousedown', (e) => {
        const pos = getMousePos(e);
        send({
            type: 'input',
            input: {
                type: 'mousedown',
                x: pos.x,
                y: pos.y,
                button: e.button
            }
        });
    });

    video.addEventListener('mouseup', (e) => {
        const pos = getMousePos(e);
        send({
            type: 'input',
            input: {
                type: 'mouseup',
                x: pos.x,
                y: pos.y,
                button: e.button
            }
        });
    });

    video.addEventListener('mousemove', (e) => {
        const pos = getMousePos(e);
        
        // Move fake cursor
        cursorEl.style.display = 'block';
        cursorEl.style.left = `${e.clientX}px`;
        cursorEl.style.top = `${e.clientY}px`;

        send({
            type: 'input',
            input: {
                type: 'mousemove',
                x: pos.x,
                y: pos.y
            }
        });
    });

    let lastWheelTime = 0;
    video.addEventListener('wheel', (e) => {
        e.preventDefault();
        const now = Date.now();
        if (now - lastWheelTime < 50) return; // Throttling
        lastWheelTime = now;

        const pos = getMousePos(e);
        send({
            type: 'input',
            input: {
                type: 'wheel',
                x: pos.x,
                y: pos.y,
                deltaX: Math.round(e.deltaX),
                deltaY: Math.round(e.deltaY)
            }
        });
    }, { passive: false });

    window.addEventListener('keydown', (e) => {
        if (e.ctrlKey || e.metaKey) return;
        send({
            type: 'input',
            input: {
                type: 'keydown',
                key: e.key
            }
        });
    });

    window.addEventListener('keyup', (e) => {
        send({
            type: 'input',
            input: {
                type: 'keyup',
                key: e.key
            }
        });
    });

    btnApply.addEventListener('click', () => {
        let width = 1280;
        let height = 720;
        
        // Update local UI immediately
        video.style.objectFit = selectScaling.value;
        
        if (selectRes.value === 'auto') {
            width = window.innerWidth;
            height = window.innerHeight;
        } else {
            const parts = selectRes.value.split('x');
            width = parseInt(parts[0]);
            height = parseInt(parts[1]);
        }

        send({
            type: 'config',
            config: {
                width: width,
                height: height,
                fps: parseInt(selectFps.value),
                bitrate: parseInt(selectBitrate.value)
            }
        });
        
        btnApply.textContent = 'RESTARTING...';
        btnApply.disabled = true;
        
        setTimeout(() => {
            window.location.reload();
        }, 3000);
    });

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

        ws.onclose = () => {
            updateStatus('● Disconnected', 'disconnected');
            if (pc) {
                pc.close();
                pc = null;
            }
            if (reconnectAttempts < maxReconnectAttempts) {
                reconnectAttempts++;
                setTimeout(connect, 2000);
            }
        };
    }

    async function initWebRTC() {
        try {
            pc = new RTCPeerConnection({
                iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
            });

            pc.ontrack = (event) => {
                if (event.track.kind === 'video') {
                    video.srcObject = event.streams[0];
                    updateStatus('● Connected', 'connected');
                    startPlayback();
                }
            };

            pc.onicecandidate = (event) => {
                if (event.candidate) {
                    send({
                        type: 'candidate',
                        candidate: event.candidate.toJSON()
                    });
                }
            };

            const offer = await pc.createOffer({ offerToReceiveVideo: true });
            await pc.setLocalDescription(offer);

            await new Promise((resolve) => {
                if (pc.iceGatheringState === 'complete') resolve();
                else pc.onicegatheringstatechange = () => { if (pc.iceGatheringState === 'complete') resolve(); };
            });

            send({ type: 'offer', sdp: pc.localDescription });
        } catch (err) {
            updateStatus('● Connection failed', 'disconnected');
        }
    }

    async function handleMessage(msg) {
        switch (msg.type) {
            case 'answer':
                if (pc && msg.sdp) await pc.setRemoteDescription(new RTCSessionDescription(msg.sdp));
                break;
            case 'candidate':
                if (pc && msg.candidate) await pc.addIceCandidate(new RTCIceCandidate(msg.candidate));
                break;
            case 'cursor':
                if (msg.cursor) {
                    cursorEl.className = '';
                }
                break;
            case 'error':
                updateStatus('● Error: ' + msg.error, 'disconnected');
                break;
        }
    }

    function send(msg) {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify(msg));
        }
    }

    connect();
})();
