class WebSocketClient {
  constructor(url) {
    this.url = url
    this.socket = null
    this.reconnectInterval = 3000
  }

  connect(onMessage) {
    this.socket = new WebSocket(this.url)

    this.socket.onopen = () => {
      console.log('WebSocket connected')
    }

    this.socket.onmessage = (event) => {
      const message = JSON.parse(event.data)
      onMessage(message)
    }

    this.socket.onclose = () => {
      console.log('WebSocket disconnected')
      setTimeout(() => this.connect(onMessage), this.reconnectInterval)
    }

    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error)
    }
  }

  send(data) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(data))
    } else {
      console.error('WebSocket not connected')
    }
  }

  close() {
    if (this.socket) {
      this.socket.close()
    }
  }
}

export default WebSocketClient