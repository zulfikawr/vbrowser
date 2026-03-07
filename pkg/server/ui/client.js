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
  const notification = document.getElementById("notification");
  const inputOverlay = document.getElementById("input-overlay");

  const controlIndicator = document.getElementById("control-indicator");
  const controlText = document.getElementById("control-text");
  const btnTakeControl = document.getElementById("btn-take-control");
  const sessionsList = document.getElementById("sessions-list");
  const sessionCountLabel = document.getElementById("session-count-label");

  // UI Elements
  const settingsBtn = document.getElementById("settings-button");
  const sidePanel = document.getElementById("side-panel");
  const closePanel = document.getElementById("close-panel");

  const connectionScreen = document.getElementById("connection-screen");
  const videoContainer = document.getElementById("video-container");

  let ws = null;
  let pc = null;
  let reconnectAttempts = 0;
  const maxReconnectAttempts = 5;
  let heartbeatInterval = null;
  let liveInterval = null;
  let isHost = false;
  let currentSessionID = null;

  // UI Interactions
  settingsBtn.addEventListener("click", () => {
    sidePanel.classList.toggle("open");
  });

  closePanel.addEventListener("click", () => {
    sidePanel.classList.remove("open");
  });

  // Disable default context menu
  window.addEventListener("contextmenu", (e) => {
    e.preventDefault();
  });

  // Close panel and focus helper when clicking
  inputOverlay.addEventListener("click", () => {
    if (sidePanel.classList.contains("open")) {
      sidePanel.classList.remove("open");
    }
    inputOverlay.focus();
  });

  btnTakeControl.addEventListener("click", () => {
    send({ type: "control", control: { action: "request" } });
  });

  function renderSessions(sessions) {
    sessionCountLabel.textContent = `Active Sessions (${sessions.length})`;
    sessionsList.innerHTML = "";

    sessions.forEach((sess) => {
      const isMe = sess.id === currentSessionID;
      const item = document.createElement("div");
      item.className = "session-item";
      
      let actions = "";
      if (isHost && !sess.is_host) {
        actions = `
          <div class="session-actions">
            <div class="action-btn" title="Give Control" onclick="requestGiveControl('${sess.id}')">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="8.5" cy="7" r="4"></circle><polyline points="17 11 19 13 23 9"></polyline></svg>
            </div>
            <div class="action-btn kick" title="Kick User" onclick="requestKick('${sess.id}')">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>
            </div>
          </div>
        `;
      }

      item.innerHTML = `
        <div class="session-info">
          <div style="display: flex; align-items: center; gap: 6px;">
            <span class="session-id">${sess.id.substring(0, 8)}${isMe ? " (You)" : ""}</span>
            ${sess.is_host ? '<span class="session-badge host">HOST</span>' : '<span class="session-badge">VIEWER</span>'}
          </div>
          <div style="font-size: 10px; opacity: 0.4;">${sess.remote}</div>
        </div>
        ${actions}
      `;
      sessionsList.appendChild(item);
    });
  }

  window.requestGiveControl = (id) => {
    send({ type: "control", control: { action: "give", target_id: id } });
  };

  window.requestKick = (id) => {
    if (confirm("Are you sure you want to kick this user?")) {
      send({ type: "control", control: { action: "kick", target_id: id } });
    }
  };

  function startPlayback() {
    video.muted = true; // ALWAYS start muted for 100% reliable autoplay

    // Ultra-low latency video settings
    video.playsInline = true;
    if ("requestVideoFrameCallback" in video) {
      video.requestVideoFrameCallback(() => {});
    }

    if (!liveInterval) {
      liveInterval = setInterval(jumpToLive, 100);
    }

    video
      .play()
      .then(() => {
        videoContainer.classList.add("visible");
        connectionScreen.classList.add("hidden");
        overlay.classList.add("visible"); 

        setTimeout(() => {
          settingsBtn.style.opacity = "1";
          setTimeout(() => {
            settingsBtn.style.opacity = "";
          }, 2000);
        }, 1000);

        inputOverlay.focus();
      })
      .catch(() => {
        connectionScreen.classList.add("hidden");
        overlay.classList.add("visible");
      });
  }

  overlay.addEventListener("click", () => {
    video.muted = false; 
    overlay.classList.remove("visible");
    video.play();
    inputOverlay.focus();
  });

  document.addEventListener("visibilitychange", () => {
    if (!document.hidden) {
      console.log("Tab visible, performing hard re-sync...");
      hardResync();
    }
  });

  window.addEventListener("focus", () => {
    console.log("Window focused, performing hard re-sync...");
    syncClipboard();
    hardResync();
  });

  function hardResync() {
    if (video.srcObject) {
      // Clear WebRTC buffer by re-attaching stream
      const stream = video.srcObject;
      video.srcObject = null;
      video.srcObject = stream;
      video.play().catch(() => {});
    }
    requestKeyframe();
    // Immediate catch-up
    setTimeout(jumpToLive, 50);
    setTimeout(jumpToLive, 150);
    setTimeout(jumpToLive, 300);
  }

  function jumpToLive() {
    if (video.buffered.length > 0) {
      const last = video.buffered.end(video.buffered.length - 1);
      const diff = last - video.currentTime;
      
      if (diff > 0.1) {
        // Very aggressive jump for any noticeable lag
        video.currentTime = last;
      } else if (diff > 0.03) {
        // Minor catch-up speed
        video.playbackRate = 1.1;
      } else {
        video.playbackRate = 1.0;
      }
    }
  }

  function requestKeyframe() {
    send({ type: "pli" });
  }

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

  inputOverlay.addEventListener("mousedown", (e) => {
    if (sidePanel.classList.contains("open") || !isHost) return;

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
    inputOverlay.focus();
  });

  inputOverlay.addEventListener("mouseup", (e) => {
    if (sidePanel.classList.contains("open") || !isHost) return;

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

  inputOverlay.addEventListener("mousemove", (e) => {
    if (sidePanel.classList.contains("open") || !isHost) return;

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

  inputOverlay.addEventListener(
    "wheel",
    (e) => {
      if (sidePanel.classList.contains("open") || !isHost) return;

      e.preventDefault();
      const pos = getMousePos(e);

      let deltaX = e.deltaX;
      let deltaY = e.deltaY;

      if (e.deltaMode === 1) {
        deltaX *= 15;
        deltaY *= 15;
      } else if (e.deltaMode === 2) {
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
    if (sidePanel.classList.contains("open") || !isHost) return true;

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
    if (sidePanel.classList.contains("open") || !isHost) return;

    send({
      type: "input",
      input: {
        type: "keyup",
        keysym: keysym,
      },
    });
  };

  keyboard.listenTo(inputOverlay);

  // Focus helper on global keydown to ensure we can catch paste
  document.addEventListener("keydown", (e) => {
    if (!sidePanel.classList.contains("open") && document.activeElement !== inputOverlay) {
      inputOverlay.focus();
    }
  });

  function showNotification(msg) {
    notification.textContent = msg;
    notification.style.opacity = "1";
    setTimeout(() => {
      notification.style.opacity = "0";
    }, 2000);
  }

  // Automatic Clipboard Sync: Local -> Remote
  inputOverlay.addEventListener("paste", (e) => {
    if (!isHost) return;
    const text = e.clipboardData.getData("text");
    if (text) {
      send({ type: "clipboard", clipboard: text });
      showNotification("Pasted to virtual browser");
      setTimeout(() => { inputOverlay.value = ""; }, 10);
    }
  });

  async function syncClipboard() {
    if (!isHost || !window.document.hasFocus()) return;
    try {
      const text = await navigator.clipboard.readText();
      if (text) {
        send({ type: "clipboard", clipboard: text });
      }
    } catch (err) {
      // Restricted
    }
  }

  inputOverlay.addEventListener("mouseenter", () => {
    syncClipboard();
  });

  btnApply.addEventListener("click", () => {
    let newWidth = 1280;
    let newHeight = 720;

    if (selectRes.value === "auto") {
      newWidth = window.innerWidth;
      newHeight = window.innerHeight;
    } else {
      const parts = selectRes.value.split("x");
      newWidth = parseInt(parts[0]);
      newHeight = parseInt(parts[1]);
    }

    video.style.objectFit = selectScaling.value;

    const resChanged =
      newWidth !== video.videoWidth || newHeight !== video.videoHeight;

    send({
      type: "config",
      config: {
        width: newWidth,
        height: newHeight,
        fps: parseInt(selectFps.value),
        bitrate: parseInt(selectBitrate.value),
      },
    });

    if (resChanged) {
      btnApply.textContent = "RESTARTING...";
      btnApply.disabled = true;
      
      const pollHealth = () => {
        fetch("/health")
          .then(res => res.json())
          .then(data => {
            if (data.status === "ok") {
              window.location.reload();
            } else {
              setTimeout(pollHealth, 1000);
            }
          })
          .catch(() => {
            setTimeout(pollHealth, 1000);
          });
      };
      
      setTimeout(pollHealth, 2000);
    } else {
      btnApply.textContent = "APPLIED";
      setTimeout(() => {
        btnApply.textContent = "Apply Changes";
        btnApply.disabled = false;
      }, 2000);
    }
  });

  function updateStatus(className) {
    if (className === "connecting") {
      // Don't hide the video during resets to avoid flicker
      if (!pc) videoContainer.classList.remove("visible");
      connectionScreen.classList.remove("hidden");
    } else if (className === "connected") {
      // Handled by startPlayback
    } else if (className === "disconnected") {
      videoContainer.classList.remove("visible");
      connectionScreen.classList.remove("hidden");
      video.srcObject = null;
    }
  }

  function disconnect() {
    if (heartbeatInterval) {
      clearInterval(heartbeatInterval);
      heartbeatInterval = null;
    }
    if (liveInterval) {
      clearInterval(liveInterval);
      liveInterval = null;
    }
    if (ws) {
      ws.onclose = null;
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
      reconnectAttempts = 0;
      initWebRTC();

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
    updateStatus("connecting");
    try {
      if (pc) {
        pc.onconnectionstatechange = null;
        pc.onicecandidate = null;
        pc.ontrack = null;
        pc.close();
      }

      pc = new RTCPeerConnection({
        iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
      });

      pc.onconnectionstatechange = () => {
        if (pc.connectionState === "failed") {
          updateStatus("disconnected");
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

      pc.ontrack = (event) => {
        if (event.track.kind === "video") {
          const stream = event.streams[0];
          video.srcObject = stream;

          const videoTrack = stream.getVideoTracks()[0];
          if (videoTrack) {
            if ("contentHint" in videoTrack) videoTrack.contentHint = "motion";
            if ("playoutDelayHint" in videoTrack) videoTrack.playoutDelayHint = 0;
          }

          const audioTrack = stream.getAudioTracks()[0];
          if (audioTrack && "playoutDelayHint" in audioTrack) {
            audioTrack.playoutDelayHint = 0;
          }

          updateStatus("connected");
          startPlayback();
        }
      };

      pc.onconnectionstatechange = () => {
        if (pc.connectionState === "failed") {
          console.error("WebRTC connection failed, reconnecting...");
          updateStatus("disconnected");
          if (ws) ws.close();
        }
      };

      const offer = await pc.createOffer({
        offerToReceiveVideo: true,
        offerToReceiveAudio: true,
      });
      await pc.setLocalDescription(offer);

      send({ type: "offer", sdp: pc.localDescription });
    } catch (err) {
      console.error("initWebRTC failed:", err);
      updateStatus("disconnected");
      if (ws) ws.close();
    }
  }

  async function handleMessage(msg) {
    switch (msg.type) {
      case "pong":
        break;
      case "answer":
        if (pc && msg.sdp)
          await pc.setRemoteDescription(new RTCSessionDescription(msg.sdp));
        break;
      case "candidate":
        if (pc && msg.candidate && msg.candidate.candidate)
          await pc.addIceCandidate(new RTCIceCandidate(msg.candidate));
        break;
      case "config":
        if (msg.session_id) {
          currentSessionID = msg.session_id;
        }
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
        if (msg.control) {
          isHost = msg.control.is_host;
          if (controlIndicator) controlIndicator.style.background = isHost ? "#22c55e" : "#ff4b4b";
          if (controlText) controlText.textContent = isHost ? "HOST" : "VIEWER";
          if (btnTakeControl) btnTakeControl.style.display = isHost ? "none" : "block";
        }
        if (msg.sessions) {
          renderSessions(msg.sessions);
        }
        break;
      case "control":
        if (msg.control) {
          isHost = msg.control.is_host;
          if (controlIndicator) controlIndicator.style.background = isHost ? "#22c55e" : "#ff4b4b";
          if (controlText) controlText.textContent = isHost ? "HOST" : "VIEWER";
          if (btnTakeControl) btnTakeControl.style.display = isHost ? "none" : "block";
        }
        if (msg.sessions) {
          renderSessions(msg.sessions);
        }
        break;
      case "clipboard":
        if (msg.clipboard && isHost) {
          navigator.clipboard
            .writeText(msg.clipboard)
            .then(() => {
              showNotification("Clipboard synced from virtual browser");
            })
            .catch(() => {});
        }
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
