import { defineStore } from 'pinia'
import api from '../services/api'
import websocket from '../services/websocket'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    accessToken: localStorage.getItem('access_token'),
    refreshToken: localStorage.getItem('refresh_token'),
    loading: false,
    error: null
  }),

  getters: {
    isAuthenticated: (state) => !!state.accessToken,
    currentUser: (state) => state.user
  },

  actions: {
    async register(email, username, password) {
      this.loading = true
      this.error = null

      try {
        const response = await api.post('/auth/register', {
          email,
          username,
          password
        })

        const { user, access_token, refresh_token } = response.data

        this.user = user
        this.accessToken = access_token
        this.refreshToken = refresh_token

        localStorage.setItem('access_token', access_token)
        localStorage.setItem('refresh_token', refresh_token)

        // WebSocket 연결은 별도로 처리 (실패해도 회원가입 성공)
        websocket.connect(access_token).catch(err => {
          console.error('WebSocket connection failed:', err)
        })

        return true
      } catch (error) {
        this.error = error.response?.data?.error || 'Registration failed'
        return false
      } finally {
        this.loading = false
      }
    },

    async login(email, password) {
      this.loading = true
      this.error = null

      try {
        const response = await api.post('/auth/login', {
          email,
          password
        })

        const { user, access_token, refresh_token } = response.data

        this.user = user
        this.accessToken = access_token
        this.refreshToken = refresh_token

        localStorage.setItem('access_token', access_token)
        localStorage.setItem('refresh_token', refresh_token)

        // WebSocket 연결은 별도로 처리 (실패해도 로그인 성공)
        websocket.connect(access_token).catch(err => {
          console.error('WebSocket connection failed:', err)
        })

        return true
      } catch (error) {
        this.error = error.response?.data?.error || 'Login failed'
        return false
      } finally {
        this.loading = false
      }
    },

    async logout() {
      try {
        await api.post('/auth/logout', {
          refresh_token: this.refreshToken
        })
      } catch (error) {
        console.error('Logout error', error)
      }

      websocket.disconnect()

      this.user = null
      this.accessToken = null
      this.refreshToken = null

      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
    },

    async fetchCurrentUser() {
      if (!this.accessToken) return

      try {
        const response = await api.get('/auth/me')
        this.user = response.data

        if (!websocket.isConnected) {
          await websocket.connect(this.accessToken)
        }
      } catch (error) {
        console.error('Failed to fetch user', error)
        if (error.response?.status === 401) {
          this.logout()
        }
      }
    },

    async initAuth() {
      if (this.accessToken) {
        await this.fetchCurrentUser()
      }
    }
  }
})
