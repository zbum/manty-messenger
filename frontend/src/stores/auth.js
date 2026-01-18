import { defineStore } from 'pinia'
import api from '../services/api'
import websocket from '../services/websocket'
import { useChatStore } from './chat'
import keycloak, { login as keycloakLogin, logout as keycloakLogout, getToken, isAuthenticated } from '../services/keycloak'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    loading: false,
    error: null
  }),

  getters: {
    isAuthenticated: () => isAuthenticated(),
    currentUser: (state) => state.user,
    accessToken: () => getToken()
  },

  actions: {
    async login() {
      return keycloakLogin()
    },

    async logout() {
      try {
        await api.post('/auth/logout')
      } catch (error) {
        console.error('Logout error', error)
      }

      // 로그아웃 시 완전 초기화
      websocket.cleanup()

      const chatStore = useChatStore()
      chatStore.reset()

      this.user = null

      return keycloakLogout()
    },

    async fetchCurrentUser() {
      if (!isAuthenticated()) return

      try {
        const response = await api.get('/auth/me')
        this.user = response.data

        if (!websocket.isConnected) {
          await websocket.connect(getToken())
        }
      } catch (error) {
        console.error('Failed to fetch user', error)
        if (error.response?.status === 401) {
          await this.logout()
        }
      }
    },

    async initAuth() {
      if (isAuthenticated()) {
        await this.fetchCurrentUser()
      }
    },

    getKeycloakInstance() {
      return keycloak
    }
  }
})
