// WebRTC client for vbrowser
(function () {
  "use strict";

  const video = document.getElementById("video");
  const overlay = document.getElementById("overlay");
  const cursorEl = document.getElementById("cursor");
  const selectRes = document.getElementById("select-res");
  const selectScaling = document.getElementById("select-scaling");
  const selectFps = document.getElementById("select-fps");
  const selectBitrate = document.getElementById("select-bitrate");
  const btnApply = document.getElementById("btn-apply");

  // UI Elements
  const settingsBtn = document.getElementById("settings-button");
  const sidePanel = document.getElementById("side-panel");
  const closePanel = document.getElementById("close-panel");

  const connectionScreen = document.getElementById("connection-screen");
  const connectionIcon = document.getElementById("connection-icon");
  const connectionText = document.getElementById("connection-text");
  const connectionSubtext = document.getElementById("connection-subtext");
  const connectionSpinner = document.getElementById("connection-spinner");

  const videoContainer = document.getElementById("video-container");

  let ws = null;
  let pc = null;
  let reconnectAttempts = 0;
  const maxReconnectAttempts = 5;
  let heartbeatInterval = null;

  // UI Interactions
  settingsBtn.addEventListener("click", () => {
    sidePanel.classList.toggle("open");
  });

  closePanel.addEventListener("click", () => {
    sidePanel.classList.remove("open");
  });

  // Close panel when clicking outside on the video
  video.addEventListener("click", () => {
    if (sidePanel.classList.contains("open")) {
      sidePanel.classList.remove("open");
    }
  });

  function startPlayback() {
    video.muted = true; // ALWAYS start muted for 100% reliable autoplay

    // Ultra-low latency video settings
    video.playsInline = true;
    if ("requestVideoFrameCallback" in video) {
      // Use requestVideoFrameCallback for frame-perfect rendering
      video.requestVideoFrameCallback(() => {});
    }

    video
      .play()
      .then(() => {
        videoContainer.classList.add("visible");
        connectionScreen.classList.add("hidden");
        overlay.classList.add("visible"); // Show play button so user can unmute

        // Brief "Tease" for settings
        setTimeout(() => {
          settingsBtn.style.opacity = "1";
          setTimeout(() => {
            settingsBtn.style.opacity = "";
          }, 2000);
        }, 1000);

        console.log("Autoplay successful (muted)");
      })
      .catch((err) => {
        console.error("Autoplay failed:", err);
        connectionScreen.classList.add("hidden");
        overlay.classList.add("visible");
      });
  }

  overlay.addEventListener("click", () => {
    video.muted = false; // Unmute on user click
    overlay.classList.remove("visible");
    video.play();
  });

  // Handle tab visibility changes to prevent throttling lag
  document.addEventListener("visibilitychange", () => {
    if (document.hidden) {
      console.log("Tab backgrounded, disconnecting to prevent lag buildup");
      disconnect();
    } else {
      console.log("Tab foregrounded, reconnecting for fresh stream");
      connect();
    }
  });

  function getMousePos(e) {
    const rect = video.getBoundingClientRect();

    if (selectScaling.value === "fill") {
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

    const x =
      (e.clientX - rect.left - offsetX) * (video.videoWidth / contentWidth);
    const y =
      (e.clientY - rect.top - offsetY) * (video.videoHeight / contentHeight);

    return { x: Math.round(x), y: Math.round(y) };
  }

  video.addEventListener("mousedown", (e) => {
    if (sidePanel.classList.contains("open")) return;

    const pos = getMousePos(e);
    send({
      type: "input",
      input: {
        type: "mousedown",
        x: pos.x,
        y: pos.y,
        button: e.button,
      },
    });
  });

  video.addEventListener("mouseup", (e) => {
    if (sidePanel.classList.contains("open")) return;

    const pos = getMousePos(e);
    send({
      type: "input",
      input: {
        type: "mouseup",
        x: pos.x,
        y: pos.y,
        button: e.button,
      },
    });
  });

  video.addEventListener("mousemove", (e) => {
    if (sidePanel.classList.contains("open")) return;

    const pos = getMousePos(e);
    send({
      type: "input",
      input: {
        type: "mousemove",
        x: pos.x,
        y: pos.y,
      },
    });
  });

  video.addEventListener(
    "wheel",
    (e) => {
      if (sidePanel.classList.contains("open")) return;

      e.preventDefault();
      const pos = getMousePos(e);

      let deltaX = e.deltaX;
      let deltaY = e.deltaY;

      if (e.deltaMode === 1) {
        // DOM_DELTA_LINE
        deltaX *= 15;
        deltaY *= 15;
      } else if (e.deltaMode === 2) {
        // DOM_DELTA_PAGE
        deltaX *= 100;
        deltaY *= 100;
      }

      deltaX = deltaX / 100;
      deltaY = deltaY / 100;

      if (Math.abs(deltaX) < 0.01 && Math.abs(deltaY) < 0.01) return;

      send({
        type: "input",
        input: {
          type: "wheel",
          x: pos.x,
          y: pos.y,
          deltaX: deltaX,
          deltaY: deltaY,
        },
      });
    },
    { passive: false },
  );

  // Use Guacamole Keyboard for proper key handling
  const keyboard = new Guacamole.Keyboard(document);

  keyboard.onkeydown = (keysym) => {
    if (sidePanel.classList.contains("open")) return true;

    send({
      type: "input",
      input: {
        type: "keydown",
        keysym: keysym,
      },
    });
    return false;
  };

  keyboard.onkeyup = (keysym) => {
    if (sidePanel.classList.contains("open")) return;

    send({
      type: "input",
      input: {
        type: "keyup",
        keysym: keysym,
      },
    });
  };

  keyboard.listenTo(document);

  btnApply.addEventListener("click", () => {
    let width = 1280;
    let height = 720;

    video.style.objectFit = selectScaling.value;

    if (selectRes.value === "auto") {
      width = window.innerWidth;
      height = window.innerHeight;
    } else {
      const parts = selectRes.value.split("x");
      width = parseInt(parts[0]);
      height = parseInt(parts[1]);
    }

    send({
      type: "config",
      config: {
        width: width,
        height: height,
        fps: parseInt(selectFps.value),
        bitrate: parseInt(selectBitrate.value),
      },
    });

    btnApply.textContent = "RESTARTING...";
    btnApply.disabled = true;

    setTimeout(() => {
      window.location.reload();
    }, 3000);
  });

  function updateStatus(className) {
    if (className === "connecting") {
      connectionScreen.classList.remove("hidden");
      connectionText.textContent = "Connecting to vbrowser...";
      connectionIcon.innerHTML = `<svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
                <line x1="8" y1="21" x2="16" y2="21"></line>
                <line x1="12" y1="17" x2="12" y2="21"></line>
            </svg>`;
      connectionSpinner.style.display = "block";
      videoContainer.classList.remove("visible");
    } else if (className === "connected") {
      // Screen is hidden by startPlayback()
    } else if (className === "disconnected") {
      connectionScreen.classList.remove("hidden");
      connectionText.textContent = "Disconnected";
      connectionSubtext.textContent = "The connection to the server was lost";
      connectionIcon.innerHTML = `<svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
                <line x1="12" y1="9" x2="12" y2="13"></line>
                <line x1="12" y1="17" x2="12.01" y2="17"></line>
            </svg>`;
      connectionSpinner.style.display = "none";
      videoContainer.classList.remove("visible");
      video.srcObject = null;
    }
  }

  function disconnect() {
    if (heartbeatInterval) {
      clearInterval(heartbeatInterval);
      heartbeatInterval = null;
    }
    if (ws) {
      ws.onclose = null; // Prevent reconnect on intentional close
      ws.close();
      ws = null;
    }
    if (pc) {
      pc.close();
      pc = null;
    }
    video.srcObject = null;
  }

  function connect() {
    disconnect();

    fetch("/health")
      .then((res) => res.json())
      .then((data) => {
        if (data.status === "ok") {
          proceedToConnect();
        } else {
          updateStatus("disconnected");
          setTimeout(connect, 2000);
        }
      })
      .catch(() => {
        updateStatus("connecting");
        setTimeout(connect, 2000);
      });
  }

  function proceedToConnect() {
    updateStatus("connecting");
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log("WebSocket connected");
      reconnectAttempts = 0;
      initWebRTC();

      // Start heartbeat
      heartbeatInterval = setInterval(() => {
        send({ type: "ping" });
      }, 5000);
    };

    ws.onmessage = async (event) => {
      try {
        const msg = JSON.parse(event.data);
        await handleMessage(msg);
      } catch (err) {
        console.error("Failed to handle message:", err);
      }
    };

    ws.onclose = () => {
      updateStatus("disconnected");
      if (reconnectAttempts < maxReconnectAttempts) {
        reconnectAttempts++;
        setTimeout(connect, 2000);
      }
    };
  }

  async function initWebRTC() {
    try {
      pc = new RTCPeerConnection({
        iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
        bundlePolicy: "max-bundle",
        rtcpMuxPolicy: "require",
        sdpSemantics: "unified-plan",
      });

      pc.ontrack = (event) => {
        if (event.track.kind === "video") {
          const stream = event.streams[0];
          video.srcObject = stream;

          const videoTrack = stream.getVideoTracks()[0];
          if (videoTrack) {
            if ("contentHint" in videoTrack) {
              videoTrack.contentHint = "motion";
            }
            if ("playoutDelayHint" in videoTrack) {
              videoTrack.playoutDelayHint = 0;
            }
          }

          updateStatus("connected");
          startPlayback();
        }
      };

      pc.onicecandidate = (event) => {
        if (event.candidate) {
          send({
            type: "candidate",
            candidate: event.candidate.toJSON(),
          });
        }
      };

      const offer = await pc.createOffer({
        offerToReceiveVideo: true,
        offerToReceiveAudio: true,
        voiceActivityDetection: false,
      });
      await pc.setLocalDescription(offer);

      await new Promise((resolve) => {
        if (pc.iceGatheringState === "complete") resolve();
        else
          pc.onicegatheringstatechange = () => {
            if (pc.iceGatheringState === "complete") resolve();
          };
      });

      send({ type: "offer", sdp: pc.localDescription });
    } catch (err) {
      updateStatus("disconnected");
    }
  }

  async function handleMessage(msg) {
    switch (msg.type) {
      case "pong":
        // Heartbeat success
        break;
      case "answer":
        if (pc && msg.sdp)
          await pc.setRemoteDescription(new RTCSessionDescription(msg.sdp));
        break;
      case "candidate":
        if (pc && msg.candidate)
          await pc.addIceCandidate(new RTCIceCandidate(msg.candidate));
        break;
      case "config":
        if (msg.config) {
          const resStr = `${msg.config.width}x${msg.config.height}`;
          let resFound = false;
          for (let i = 0; i < selectRes.options.length; i++) {
            if (selectRes.options[i].value === resStr) {
              selectRes.selectedIndex = i;
              resFound = true;
              break;
            }
          }
          if (!resFound && msg.config.width > 0) {
            const opt = document.createElement("option");
            opt.value = resStr;
            opt.text = `${resStr} (Current)`;
            selectRes.add(opt);
            selectRes.value = resStr;
          }
          selectFps.value = msg.config.fps.toString();
          selectBitrate.value = msg.config.bitrate.toString();
        }
        break;
      case "error":
        updateStatus("disconnected");
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
