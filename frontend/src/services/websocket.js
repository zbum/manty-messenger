class WebSocketService {
  constructor() {
    this.socket = null
    this.listeners = new Map()
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
    this.reconnectDelay = 1000
    this.isConnecting = false
    this.currentRoomId = null
    this.pendingMessages = []
  }

  connect(token) {
    if (this.isConnecting || (this.socket && this.socket.readyState === WebSocket.OPEN)) {
      return Promise.resolve()
    }

    this.isConnecting = true

    return new Promise((resolve, reject) => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const host = window.location.host
      const wsUrl = `${protocol}//${host}/messenger/ws?token=${token}`

      this.socket = new WebSocket(wsUrl)

      this.socket.onopen = () => {
        console.log('WebSocket connected')
        const wasReconnecting = this.reconnectAttempts > 0
        this.reconnectAttempts = 0
        this.isConnecting = false

        // Send pending messages (for initial connection)
        while (this.pendingMessages.length > 0) {
          const msg = this.pendingMessages.shift()
          console.log('Sending pending message:', msg.type)
          this.socket.send(JSON.stringify(msg))
        }

        // Rejoin room only if reconnecting (not initial connection)
        if (wasReconnecting && this.currentRoomId) {
          console.log('Rejoining room after reconnect:', this.currentRoomId)
          this.send('join_room', { room_id: this.currentRoomId })
        }

        resolve()
      }

      this.socket.onclose = (event) => {
        console.log('WebSocket closed', event.code, event.reason)
        this.isConnecting = false
        this.handleReconnect(token)
      }

      this.socket.onerror = (error) => {
        console.error('WebSocket error', error)
        this.isConnecting = false
        reject(error)
      }

      this.socket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)
          this.handleMessage(message)
        } catch (e) {
          console.error('Failed to parse message', e)
        }
      }
    })
  }

  handleMessage(message) {
    const listeners = this.listeners.get(message.type) || []
    listeners.forEach(callback => callback(message.payload, message))
  }

  on(type, callback) {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, [])
    }
    this.listeners.get(type).push(callback)

    return () => this.off(type, callback)
  }

  off(type, callback) {
    const listeners = this.listeners.get(type) || []
    const index = listeners.indexOf(callback)
    if (index > -1) {
      listeners.splice(index, 1)
    }
  }

  send(type, payload = {}) {
    const message = {
      type,
      payload,
      timestamp: new Date().toISOString(),
      request_id: this.generateId()
    }

    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(message))
    } else if (this.isConnecting || this.socket) {
      // Queue message if connecting or socket exists but not ready
      console.log('Queuing message:', type)
      this.pendingMessages.push(message)
    }
  }

  joinRoom(roomId) {
    this.currentRoomId = roomId
    this.send('join_room', { room_id: roomId })
  }

  leaveRoom(roomId) {
    if (this.currentRoomId === roomId) {
      this.currentRoomId = null
    }
    this.send('leave_room', { room_id: roomId })
  }

  sendMessage(roomId, content, messageType = 'text') {
    this.send('send_message', {
      room_id: roomId,
      content,
      message_type: messageType
    })
  }

  setTyping(roomId, isTyping) {
    this.send('typing', { room_id: roomId, is_typing: isTyping })
  }

  markRead(roomId, messageId) {
    this.send('mark_read', { room_id: roomId, message_id: messageId })
  }

  handleReconnect(token) {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)
      console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`)
      setTimeout(() => this.connect(token), delay)
    } else {
      console.error('Max reconnection attempts reached')
    }
  }

  disconnect() {
    if (this.socket) {
      this.socket.close()
      this.socket = null
    }
    this.listeners.clear()
    this.reconnectAttempts = 0
    this.currentRoomId = null
    this.pendingMessages = []
  }

  generateId() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
      const r = Math.random() * 16 | 0
      const v = c === 'x' ? r : (r & 0x3 | 0x8)
      return v.toString(16)
    })
  }

  get isConnected() {
    return this.socket?.readyState === WebSocket.OPEN
  }
}

export default new WebSocketService()
