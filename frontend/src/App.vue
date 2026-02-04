<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from './stores/auth'
import { useChatStore } from './stores/chat'
import notificationService from './services/notification'
import InstallPrompt from './components/InstallPrompt.vue'

const authStore = useAuthStore()
const chatStore = useChatStore()
const isOffline = ref(!navigator.onLine)

function handleOnline() { isOffline.value = false }
function handleOffline() { isOffline.value = true }

onMounted(async () => {
  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)

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

onUnmounted(() => {
  window.removeEventListener('online', handleOnline)
  window.removeEventListener('offline', handleOffline)
})
</script>

<template>
  <div v-if="isOffline" class="offline-bar">
    오프라인 상태입니다. 일부 기능이 제한될 수 있습니다.
  </div>
  <router-view />
  <InstallPrompt />
</template>

<style>
#app {
  min-height: 100vh;
}

.offline-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  background: #ff9800;
  color: white;
  text-align: center;
  padding: 8px;
  font-size: 13px;
  font-weight: 500;
  z-index: 10000;
}
</style>
