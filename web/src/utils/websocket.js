class WebSocketClient {
  constructor(url) {
    this.url = url;
    this.socket = null;
    this.reconnectInterval = 3000;
    this.reconnectTimer = null;
    this.manualClosed = false;
  }

  clearReconnectTimer() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  scheduleReconnect(onMessage) {
    if (this.manualClosed) return;
    this.clearReconnectTimer();
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      if (this.manualClosed) return;
      this.connect(onMessage);
    }, this.reconnectInterval);
  }

  connect(onMessage) {
    this.manualClosed = false;
    this.clearReconnectTimer();

    this.socket = new WebSocket(this.url);

    this.socket.onopen = () => {};

    this.socket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      onMessage(message);
    };

    this.socket.onclose = () => {
      this.scheduleReconnect(onMessage);
    };

    this.socket.onerror = (error) => {
      console.error("WebSocket error:", error);
    };
  }

  send(data) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(data));
    } else {
      console.error("WebSocket not connected");
    }
  }

  close() {
    this.manualClosed = true;
    this.clearReconnectTimer();
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }
}

export default WebSocketClient;
