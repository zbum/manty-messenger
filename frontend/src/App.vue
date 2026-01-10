<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from './stores/auth'
import { useChatStore } from './stores/chat'

const authStore = useAuthStore()
const chatStore = useChatStore()

onMounted(async () => {
  await authStore.initAuth()
  if (authStore.isAuthenticated) {
    chatStore.initWebSocketListeners()
  }
})
</script>

<template>
  <router-view />
</template>

<style>
#app {
  min-height: 100vh;
}
</style>
