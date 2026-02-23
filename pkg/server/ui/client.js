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
    
    const connectionScreen = document.getElementById('connection-screen');
    const connectionIcon = document.getElementById('connection-icon');
    const connectionText = document.getElementById('connection-text');
    const connectionSubtext = document.getElementById('connection-subtext');
    const connectionSpinner = document.getElementById('connection-spinner');
    
    const videoContainer = document.getElementById('video-container');
    
    function startPlayback() {
        video.muted = true; // ALWAYS start muted for 100% reliable autoplay
        
        video.play().then(() => {
            videoContainer.classList.add('visible');
            connectionScreen.classList.add('hidden');
            overlay.classList.add('visible'); // Show play button so user can unmute
            console.log('Autoplay successful (muted)');
        }).catch(err => {
            console.error('Autoplay failed:', err);
            connectionScreen.classList.add('hidden');
            overlay.classList.add('visible');
        });
    }

    overlay.addEventListener('click', () => {
        video.muted = false; // Unmute on user click
        overlay.classList.remove('visible');
        video.play();
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
        // Only block browser shortcuts if we want to handle them in the virtual browser
        // For example, let's allow Ctrl+A, Ctrl+C, Ctrl+V, etc.
        const isModifier = e.ctrlKey || e.metaKey || e.altKey || e.shiftKey;
        
        // Prevent default for space to avoid page scrolling
        if (e.key === ' ') e.preventDefault();

        send({
            type: 'input',
            input: {
                type: 'keydown',
                key: e.key,
                ctrl: e.ctrlKey,
                alt: e.altKey,
                shift: e.shiftKey,
                meta: e.metaKey
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
        status.textContent = '• ' + message;
        status.className = className;
        
        if (className === 'connecting') {
            connectionScreen.classList.remove('hidden');
            connectionText.textContent = 'Connecting to vbrowser...';
            connectionIcon.innerHTML = `<svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
                <line x1="8" y1="21" x2="16" y2="21"></line>
                <line x1="12" y1="17" x2="12" y2="21"></line>
            </svg>`;
            connectionSpinner.style.display = 'block';
            video.classList.remove('visible');
        } else if (className === 'connected') {
            // Screen is hidden by startPlayback()
        } else if (className === 'disconnected') {
            connectionScreen.classList.remove('hidden');
            connectionText.textContent = 'Disconnected';
            connectionSubtext.textContent = 'The connection to the server was lost';
            connectionIcon.innerHTML = `<svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
                <line x1="12" y1="9" x2="12" y2="13"></line>
                <line x1="12" y1="17" x2="12.01" y2="17"></line>
            </svg>`;
            connectionSpinner.style.display = 'none';
            videoContainer.classList.remove('visible');
            
            // Clear video src to ensure it doesn't show a frozen frame
            video.srcObject = null;
        }
    }

    function connect() {
        // Health check before connecting
        fetch('/health')
            .then(res => res.json())
            .then(data => {
                if (data.status === 'ok') {
                    proceedToConnect();
                } else {
                    updateStatus('Server Unhealthy', 'disconnected');
                    setTimeout(connect, 2000);
                }
            })
            .catch(() => {
                updateStatus('Waiting for Server...', 'connecting');
                setTimeout(connect, 2000);
            });
    }

    function proceedToConnect() {
        updateStatus('Connecting...', 'connecting');
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
            updateStatus('Disconnected', 'disconnected');
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
                    updateStatus('Connected', 'connected');
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
            case 'config':
                if (msg.config) {
                    console.log('Syncing UI with server config:', msg.config);
                    // Update dropdowns to match server state
                    const resStr = `${msg.config.width}x${msg.config.height}`;
                    
                    // Check if resolution exists in options, otherwise add it or keep "auto"
                    let resFound = false;
                    for (let i = 0; i < selectRes.options.length; i++) {
                        if (selectRes.options[i].value === resStr) {
                            selectRes.selectedIndex = i;
                            resFound = true;
                            break;
                        }
                    }
                    if (!resFound && msg.config.width > 0) {
                        // If it's a custom resolution not in list, we'll just show it
                        const opt = document.createElement('option');
                        opt.value = resStr;
                        opt.text = `${resStr} (Current)`;
                        selectRes.add(opt);
                        selectRes.value = resStr;
                    }

                    selectFps.value = msg.config.fps.toString();
                    selectBitrate.value = msg.config.bitrate.toString();
                }
                break;
            case 'cursor':
                if (msg.cursor) {
                    cursorEl.className = '';
                }
                break;
            case 'error':
                updateStatus('Error: ' + msg.error, 'disconnected');
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
