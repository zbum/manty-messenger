<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from './stores/auth'
import { useChatStore } from './stores/chat'
import notificationService from './services/notification'

const authStore = useAuthStore()
const chatStore = useChatStore()

onMounted(async () => {
  await authStore.initAuth()
  if (authStore.isAuthenticated) {
    chatStore.initWebSocketListeners()
    // 알림 권한 요청
    if (notificationService.checkPermission() === 'default') {
      await notificationService.requestPermission()
    }
    // 서비스 워커 등록 및 푸시 알림 구독
    if (notificationService.checkPermission() === 'granted') {
      await notificationService.registerServiceWorker()
      await notificationService.subscribePush()
    }
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
