import { defineStore } from 'pinia'
import api from '../services/api'
import websocket from '../services/websocket'

export const useChatStore = defineStore('chat', {
  state: () => ({
    rooms: [],
    currentRoom: null,
    messages: {},
    typingUsers: {},
    loading: false,
    error: null
  }),

  getters: {
    currentMessages: (state) => {
      if (!state.currentRoom) return []
      return state.messages[state.currentRoom.id] || []
    },
    currentTypingUsers: (state) => {
      if (!state.currentRoom) return []
      return state.typingUsers[state.currentRoom.id] || []
    }
  },

  actions: {
    initWebSocketListeners() {
      websocket.on('new_message', (payload) => {
        this.addMessage(payload.room_id, {
          id: payload.id,
          room_id: payload.room_id,
          sender: payload.sender,
          content: payload.content,
          message_type: payload.message_type,
          created_at: payload.created_at
        })
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
        }
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
        } else {
          this.messages[roomId] = [...messages, ...this.messages[roomId]]
        }
      } catch (error) {
        console.error('Failed to fetch messages', error)
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

      if (!this.messages[room.id]) {
        await this.fetchMessages(room.id)
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

    addMessage(roomId, message) {
      if (!this.messages[roomId]) {
        this.messages[roomId] = []
      }
      this.messages[roomId].push(message)
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
    }
  }
})
