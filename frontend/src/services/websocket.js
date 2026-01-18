import { getValidToken } from './keycloak'

// 연결 상태 상수
const ConnectionState = {
  DISCONNECTED: 'disconnected',
  CONNECTING: 'connecting',
  CONNECTED: 'connected',
  RECONNECTING: 'reconnecting'
}

const OFFLINE_QUEUE_KEY = 'websocket_offline_queue'

class WebSocketService {
  constructor() {
    this.socket = null
    this.listeners = new Map()
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = Infinity // 무한 재연결
    this.reconnectDelay = 1000
    this.maxReconnectDelay = 30000 // 최대 30초
    this.isConnecting = false
    this.currentRoomId = null
    this.pendingMessages = []
    this.intentionalDisconnect = false
    this.currentToken = null
    this.connectionState = ConnectionState.DISCONNECTED
    this.connectionStateListeners = []

    // Heartbeat 설정
    this.heartbeatInterval = null
    this.heartbeatTimeout = null
    this.heartbeatIntervalMs = 25000 // 25초마다 ping
    this.heartbeatTimeoutMs = 10000 // 10초 내 pong 없으면 재연결

    // 오프라인 큐를 localStorage에서 복구
    this.offlineQueue = this.loadOfflineQueue()
  }

  // 오프라인 큐 localStorage에서 로드
  loadOfflineQueue() {
    try {
      const saved = localStorage.getItem(OFFLINE_QUEUE_KEY)
      return saved ? JSON.parse(saved) : []
    } catch (e) {
      console.error('Failed to load offline queue', e)
      return []
    }
  }

  // 오프라인 큐 localStorage에 저장
  saveOfflineQueue() {
    try {
      localStorage.setItem(OFFLINE_QUEUE_KEY, JSON.stringify(this.offlineQueue))
    } catch (e) {
      console.error('Failed to save offline queue', e)
    }
  }

  // 연결 상태 변경 알림
  setConnectionState(state) {
    if (this.connectionState !== state) {
      this.connectionState = state
      this.connectionStateListeners.forEach(listener => listener(state))
    }
  }

  // 연결 상태 변경 리스너 등록
  onConnectionStateChange(callback) {
    this.connectionStateListeners.push(callback)
    // 현재 상태 즉시 전달
    callback(this.connectionState)
    return () => {
      const index = this.connectionStateListeners.indexOf(callback)
      if (index > -1) {
        this.connectionStateListeners.splice(index, 1)
      }
    }
  }

  connect(token) {
    // If connecting with a different token, disconnect first
    if (this.currentToken && this.currentToken !== token && this.socket) {
      this.disconnect()
    }

    if (this.isConnecting || (this.socket && this.socket.readyState === WebSocket.OPEN)) {
      return Promise.resolve()
    }

    this.intentionalDisconnect = false
    this.currentToken = token

    this.isConnecting = true
    const wasReconnecting = this.reconnectAttempts > 0
    this.setConnectionState(wasReconnecting ? ConnectionState.RECONNECTING : ConnectionState.CONNECTING)

    // 페이지 가시성/온라인 핸들러 설정
    this.setupVisibilityHandler()

    return new Promise((resolve, reject) => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const host = window.location.host
      const wsUrl = `${protocol}//${host}/messenger/ws?token=${token}`

      this.socket = new WebSocket(wsUrl)

      this.socket.onopen = () => {
        console.log('WebSocket connected')
        const wasReconnectingNow = this.reconnectAttempts > 0
        this.reconnectAttempts = 0
        this.isConnecting = false
        this.setConnectionState(ConnectionState.CONNECTED)

        // Heartbeat 시작
        this.startHeartbeat()

        // Send pending messages (for initial connection)
        while (this.pendingMessages.length > 0) {
          const msg = this.pendingMessages.shift()
          console.log('Sending pending message:', msg.type)
          this.socket.send(JSON.stringify(msg))
        }

        // Send offline queued messages (from localStorage)
        this.flushOfflineQueue()

        // Rejoin room only if reconnecting (not initial connection)
        if (wasReconnectingNow && this.currentRoomId) {
          console.log('Rejoining room after reconnect:', this.currentRoomId)
          this.send('join_room', { room_id: this.currentRoomId })
        }

        resolve()
      }

      this.socket.onclose = (event) => {
        console.log('WebSocket closed', event.code, event.reason)
        this.isConnecting = false
        this.stopHeartbeat()
        // Only reconnect if not intentionally disconnected (e.g., logout)
        if (!this.intentionalDisconnect) {
          this.setConnectionState(ConnectionState.RECONNECTING)
          this.handleReconnect()
        } else {
          this.setConnectionState(ConnectionState.DISCONNECTED)
        }
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
    // pong 메시지는 heartbeat 타이머 리셋
    if (message.type === 'pong') {
      this.handlePong()
      return
    }
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

  sendMessage(roomId, content, messageType = 'text', fileUrl = null, thumbnailUrl = null) {
    const payload = {
      room_id: roomId,
      content,
      message_type: messageType
    }
    if (fileUrl) {
      payload.file_url = fileUrl
    }
    if (thumbnailUrl) {
      payload.thumbnail_url = thumbnailUrl
    }

    // 연결이 안 되어 있으면 오프라인 큐에 저장
    if (!this.isConnected && !this.isConnecting) {
      const message = {
        type: 'send_message',
        payload,
        timestamp: new Date().toISOString(),
        request_id: this.generateId()
      }
      console.log('Saving message to offline queue:', message.type)
      this.offlineQueue.push(message)
      this.saveOfflineQueue()
      return
    }

    this.send('send_message', payload)
  }

  // 오프라인 큐의 메시지들을 전송
  flushOfflineQueue() {
    if (this.offlineQueue.length === 0) return

    console.log(`Sending ${this.offlineQueue.length} offline queued messages`)
    while (this.offlineQueue.length > 0) {
      const msg = this.offlineQueue.shift()
      console.log('Sending offline queued message:', msg.type)
      this.socket.send(JSON.stringify(msg))
    }
    this.saveOfflineQueue()
  }

  // 오프라인 큐 개수 반환
  getOfflineQueueCount() {
    return this.offlineQueue.length
  }

  setTyping(roomId, isTyping) {
    this.send('typing', { room_id: roomId, is_typing: isTyping })
  }

  markRead(roomId, messageId) {
    this.send('mark_read', { room_id: roomId, message_id: messageId })
  }

  async handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      // exponential backoff with max delay cap
      const delay = Math.min(
        this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
        this.maxReconnectDelay
      )
      console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`)
      this.reconnectTimeout = setTimeout(async () => {
        // 재연결 시 항상 새로운 토큰을 가져옴
        const freshToken = await getValidToken()
        if (freshToken) {
          console.log('Using fresh token for reconnection')
          this.connect(freshToken)
        } else {
          console.error('Failed to get fresh token for reconnection')
          // 토큰을 가져올 수 없으면 재연결 시도
          this.handleReconnect()
        }
      }, delay)
    } else {
      console.error('Max reconnection attempts reached')
      this.setConnectionState(ConnectionState.DISCONNECTED)
    }
  }

  // Heartbeat 시작
  startHeartbeat() {
    this.stopHeartbeat()
    this.heartbeatInterval = setInterval(() => {
      if (this.socket?.readyState === WebSocket.OPEN) {
        this.send('ping', {})
        // pong 응답 대기 타이머
        this.heartbeatTimeout = setTimeout(() => {
          console.warn('Heartbeat timeout, closing connection')
          if (this.socket) {
            this.socket.close()
          }
        }, this.heartbeatTimeoutMs)
      }
    }, this.heartbeatIntervalMs)
  }

  // Heartbeat 정지
  stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval)
      this.heartbeatInterval = null
    }
    if (this.heartbeatTimeout) {
      clearTimeout(this.heartbeatTimeout)
      this.heartbeatTimeout = null
    }
  }

  // Pong 수신 처리
  handlePong() {
    if (this.heartbeatTimeout) {
      clearTimeout(this.heartbeatTimeout)
      this.heartbeatTimeout = null
    }
  }

  // 수동 재연결
  async reconnect() {
    if (!this.isConnecting && !this.isConnected) {
      console.log('Manual reconnect triggered')
      this.reconnectAttempts = 0
      if (this.reconnectTimeout) {
        clearTimeout(this.reconnectTimeout)
        this.reconnectTimeout = null
      }
      // 새로운 토큰을 가져와서 재연결
      const freshToken = await getValidToken()
      if (freshToken) {
        console.log('Using fresh token for manual reconnection')
        this.connect(freshToken)
      } else if (this.currentToken) {
        // 새 토큰을 가져올 수 없으면 기존 토큰으로 시도
        console.log('Using existing token for manual reconnection')
        this.connect(this.currentToken)
      } else {
        console.error('No token available for reconnection')
      }
    }
  }

  // 페이지 가시성 변경 시 재연결 시도
  setupVisibilityHandler() {
    if (this.visibilityHandlerSetup) return
    this.visibilityHandlerSetup = true

    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'visible') {
        // 페이지가 다시 보이면 연결 상태 확인 후 재연결
        if (!this.isConnected && !this.isConnecting) {
          console.log('Page became visible, attempting reconnect')
          this.reconnect()
        }
      }
    })

    // 온라인 상태로 돌아오면 재연결
    window.addEventListener('online', () => {
      if (!this.isConnected && !this.isConnecting) {
        console.log('Network online, attempting reconnect')
        this.reconnect()
      }
    })
  }

  disconnect() {
    this.intentionalDisconnect = true
    this.stopHeartbeat()
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
      this.reconnectTimeout = null
    }
    if (this.socket) {
      this.socket.close()
      this.socket = null
    }
    this.listeners.clear()
    this.reconnectAttempts = 0
    this.currentRoomId = null
    this.pendingMessages = []
    this.currentToken = null
    this.setConnectionState(ConnectionState.DISCONNECTED)
    this.connectionStateListeners = []
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

const websocketService = new WebSocketService()

export { ConnectionState }
export default websocketService
