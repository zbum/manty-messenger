import { defineStore } from 'pinia'
import api from '../services/api'
import websocket, { ConnectionState } from '../services/websocket'
import { useAuthStore } from './auth'
import notificationService from '../services/notification'
import { login as keycloakLogin } from '../services/keycloak'

export const useChatStore = defineStore('chat', {
  state: () => ({
    rooms: [],
    currentRoom: null,
    messages: {},
    typingUsers: {},
    loading: false,
    error: null,
    // 웹소켓 연결 상태
    connectionState: ConnectionState.DISCONNECTED,
    offlineQueueCount: 0,
    // 무한 스크롤 관련
    loadingMore: false,
    hasMore: {} // roomId별로 더 불러올 메시지가 있는지
  }),

  getters: {
    currentMessages: (state) => {
      if (!state.currentRoom) return []
      return state.messages[state.currentRoom.id] || []
    },
    currentTypingUsers: (state) => {
      if (!state.currentRoom) return []
      return state.typingUsers[state.currentRoom.id] || []
    },
    isConnected: (state) => state.connectionState === ConnectionState.CONNECTED,
    isConnecting: (state) => state.connectionState === ConnectionState.CONNECTING,
    isReconnecting: (state) => state.connectionState === ConnectionState.RECONNECTING,
    isDisconnected: (state) => state.connectionState === ConnectionState.DISCONNECTED
  },

  actions: {
    initWebSocketListeners() {
      // 연결 상태 변경 리스너 등록
      let previousState = null
      websocket.onConnectionStateChange((state) => {
        const wasDisconnected = previousState === ConnectionState.RECONNECTING ||
                                previousState === ConnectionState.DISCONNECTED
        const isNowConnected = state === ConnectionState.CONNECTED

        this.connectionState = state
        this.offlineQueueCount = websocket.getOfflineQueueCount()

        // 재연결 완료 시 현재 방의 놓친 메시지 가져오기
        if (wasDisconnected && isNowConnected && this.currentRoom) {
          console.log('Reconnected, fetching missed messages for room:', this.currentRoom.id)
          this.fetchMissedMessages(this.currentRoom.id)
        }

        previousState = state
      })

      websocket.on('new_message', (payload) => {
        const message = {
          id: payload.id,
          room_id: payload.room_id,
          sender: payload.sender,
          content: payload.content,
          message_type: payload.message_type,
          file_url: payload.file_url,
          thumbnail_url: payload.thumbnail_url,
          created_at: payload.created_at,
          unread_count: payload.unread_count
        }
        this.addMessage(payload.room_id, message)

        // 브라우저 알림 표시 (현재 보고 있는 방이 아니고, 내가 보낸 메시지가 아닌 경우)
        const authStore = useAuthStore()
        const isCurrentRoom = this.currentRoom?.id === payload.room_id
        const isMyMessage = payload.sender?.id === authStore.user?.id

        if (!isCurrentRoom && !isMyMessage) {
          const room = this.rooms.find(r => r.id === payload.room_id)
          if (room) {
            notificationService.showNewMessage(message, room, () => {
              // 알림 클릭 시 해당 채팅방으로 이동
              this.joinRoom(room)
            })
          }
        }
      })

      websocket.on('message_read', (payload) => {
        this.decreaseUnreadCount(payload.room_id)
      })

      websocket.on('user_typing', (payload) => {
        this.setUserTyping(payload.room_id, payload.user_id, payload.username, payload.is_typing)
      })

      websocket.on('user_joined', (payload) => {
        console.log('User joined room', payload)
        this.updateRoomMemberCount(payload.room_id, payload.member_count)
      })

      websocket.on('user_left', (payload) => {
        console.log('User left room', payload)
        this.updateRoomMemberCount(payload.room_id, payload.member_count)
      })

      websocket.on('presence_update', (payload) => {
        console.log('Presence update', payload)
      })

      websocket.on('room_invited', (payload) => {
        console.log('Room invited', payload)
        if (payload.room) {
          this.addRoom(payload.room)
          // 채팅방 초대 알림 표시
          notificationService.showRoomInvite(payload.room, () => {
            this.joinRoom(payload.room)
          })
        }
      })

      // 인증이 필요할 때 (토큰 만료 등)
      websocket.on('auth_required', () => {
        console.log('Authentication required, redirecting to login')
        keycloakLogin()
      })
    },

    async fetchRooms() {
      this.loading = true
      try {
        const response = await api.get('/rooms')
        this.rooms = response.data || []
      } catch (error) {
        this.error = error.response?.data?.error || 'Failed to fetch rooms'
      } finally {
        this.loading = false
      }
    },

    async fetchMessages(roomId, limit = 50, offset = 0) {
      try {
        const response = await api.get(`/rooms/${roomId}/messages`, {
          params: { limit, offset }
        })
        const messages = response.data || []

        if (!this.messages[roomId]) {
          this.messages[roomId] = []
        }

        if (offset === 0) {
          this.messages[roomId] = messages
          // 초기 로드 시 더 불러올 메시지가 있는지 확인
          this.hasMore[roomId] = messages.length >= limit
        } else {
          this.messages[roomId] = [...messages, ...this.messages[roomId]]
        }
      } catch (error) {
        console.error('Failed to fetch messages', error)
      }
    },

    // 이전 메시지 더 불러오기 (무한 스크롤)
    async loadMoreMessages(roomId) {
      if (this.loadingMore || !this.hasMore[roomId]) {
        return false
      }

      const currentMessages = this.messages[roomId] || []
      if (currentMessages.length === 0) {
        return false
      }

      this.loadingMore = true

      try {
        const limit = 50
        const offset = currentMessages.length

        const response = await api.get(`/rooms/${roomId}/messages`, {
          params: { limit, offset }
        })
        const olderMessages = response.data || []

        if (olderMessages.length > 0) {
          // 이전 메시지를 앞에 추가
          this.messages[roomId] = [...olderMessages, ...currentMessages]
        }

        // 더 불러올 메시지가 있는지 확인
        this.hasMore[roomId] = olderMessages.length >= limit

        return olderMessages.length > 0
      } catch (error) {
        console.error('Failed to load more messages', error)
        return false
      } finally {
        this.loadingMore = false
      }
    },

    // 재연결 시 놓친 메시지 가져오기
    async fetchMissedMessages(roomId) {
      try {
        const existingMessages = this.messages[roomId] || []
        const lastMessageId = existingMessages.length > 0
          ? Math.max(...existingMessages.map(m => m.id))
          : 0

        // after_id를 사용해서 놓친 메시지만 가져오기
        const response = await api.get(`/rooms/${roomId}/messages`, {
          params: { after_id: lastMessageId, limit: 100 }
        })
        const missedMessages = response.data || []

        if (missedMessages.length > 0) {
          console.log(`Found ${missedMessages.length} missed messages`)
          // 기존 메시지 목록에 추가
          this.messages[roomId] = [...existingMessages, ...missedMessages]

          // 마지막 메시지 읽음 처리
          const lastMessage = missedMessages[missedMessages.length - 1]
          this.markRead(roomId, lastMessage.id)
        }
      } catch (error) {
        console.error('Failed to fetch missed messages', error)
      }
    },

    async createRoom(name, description = '', roomType = 'group', memberIds = []) {
      try {
        const response = await api.post('/rooms', {
          name,
          description,
          room_type: roomType,
          member_ids: memberIds
        })
        this.rooms.unshift(response.data)
        return response.data
      } catch (error) {
        this.error = error.response?.data?.error || 'Failed to create room'
        return null
      }
    },

    async joinRoom(room) {
      if (this.currentRoom?.id === room.id) return

      if (this.currentRoom) {
        websocket.leaveRoom(this.currentRoom.id)
      }

      this.currentRoom = room
      websocket.joinRoom(room.id)

      // Save to localStorage for persistence
      localStorage.setItem('currentRoomId', room.id.toString())

      if (!this.messages[room.id]) {
        await this.fetchMessages(room.id)
      }

      // Mark messages as read
      const messages = this.messages[room.id]
      if (messages && messages.length > 0) {
        const lastMessage = messages[messages.length - 1]
        this.markRead(room.id, lastMessage.id)
      }
    },

    async restoreLastRoom() {
      const savedRoomId = localStorage.getItem('currentRoomId')
      if (savedRoomId && this.rooms.length > 0) {
        const room = this.rooms.find(r => r.id === parseInt(savedRoomId))
        if (room) {
          await this.joinRoom(room)
        }
      }
    },

    leaveCurrentRoom() {
      if (this.currentRoom) {
        websocket.leaveRoom(this.currentRoom.id)
        this.currentRoom = null
      }
    },

    sendMessage(content) {
      if (!this.currentRoom || !content.trim()) return

      websocket.sendMessage(this.currentRoom.id, content.trim())
    },

    sendFileMessage(content, messageType, fileUrl, thumbnailUrl) {
      if (!this.currentRoom || !fileUrl) return

      websocket.sendMessage(this.currentRoom.id, content, messageType, fileUrl, thumbnailUrl)
    },

    sendStickerMessage(stickerId) {
      if (!this.currentRoom || !stickerId) return

      websocket.sendMessage(this.currentRoom.id, stickerId, 'sticker')
    },

    addMessage(roomId, message) {
      if (!this.messages[roomId]) {
        this.messages[roomId] = []
      }
      this.messages[roomId].push(message)

      // If user is viewing this room AND not the sender, mark as read
      const authStore = useAuthStore()
      if (this.currentRoom?.id === roomId && message.sender?.id !== authStore.user?.id) {
        this.markRead(roomId, message.id)
      }
    },

    setTyping(isTyping) {
      if (!this.currentRoom) return
      websocket.setTyping(this.currentRoom.id, isTyping)
    },

    setUserTyping(roomId, userId, username, isTyping) {
      if (!this.typingUsers[roomId]) {
        this.typingUsers[roomId] = []
      }

      const index = this.typingUsers[roomId].findIndex(u => u.userId === userId)

      if (isTyping && index === -1) {
        this.typingUsers[roomId].push({ userId, username })
      } else if (!isTyping && index !== -1) {
        this.typingUsers[roomId].splice(index, 1)
      }
    },

    async searchUsers(query) {
      try {
        const response = await api.get('/users', { params: { q: query } })
        return response.data || []
      } catch (error) {
        console.error('Failed to search users', error)
        return []
      }
    },

    async leaveRoom(roomId) {
      try {
        await api.post(`/rooms/${roomId}/leave`)
        this.rooms = this.rooms.filter(r => r.id !== roomId)
        if (this.currentRoom?.id === roomId) {
          this.currentRoom = null
        }
      } catch (error) {
        this.error = error.response?.data?.error || 'Failed to leave room'
      }
    },

    updateRoomMemberCount(roomId, memberCount) {
      const room = this.rooms.find(r => r.id === roomId)
      if (room) {
        room.member_count = memberCount
      }
      if (this.currentRoom?.id === roomId) {
        this.currentRoom.member_count = memberCount
      }
    },

    addRoom(room) {
      // Check if room already exists
      const existingRoom = this.rooms.find(r => r.id === room.id)
      if (!existingRoom) {
        this.rooms.unshift(room)
      }
    },

    decreaseUnreadCount(roomId) {
      const messages = this.messages[roomId]
      if (messages) {
        messages.forEach(msg => {
          if (msg.unread_count > 0) {
            msg.unread_count--
          }
        })
      }
    },

    markRead(roomId, messageId) {
      if (!roomId) return
      websocket.markRead(roomId, messageId)
      // Decrease unread count locally since broadcast excludes self
      this.decreaseUnreadCount(roomId)
    },

    // 수동 재연결
    reconnect() {
      websocket.reconnect()
    },

    reset() {
      this.rooms = []
      this.currentRoom = null
      this.messages = {}
      this.typingUsers = {}
      this.loading = false
      this.error = null
      this.connectionState = ConnectionState.DISCONNECTED
      this.offlineQueueCount = 0
      this.loadingMore = false
      this.hasMore = {}
      localStorage.removeItem('currentRoomId')
    }
  }
})
