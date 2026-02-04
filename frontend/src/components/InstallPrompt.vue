<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const DISMISS_KEY = 'pwa-install-dismissed'
const show = ref(false)
let deferredPrompt = null

function handleBeforeInstallPrompt(e) {
  e.preventDefault()
  deferredPrompt = e
  const dismissed = localStorage.getItem(DISMISS_KEY)
  if (!dismissed) {
    show.value = true
  }
}

async function install() {
  if (!deferredPrompt) return
  deferredPrompt.prompt()
  const { outcome } = await deferredPrompt.userChoice
  if (outcome === 'accepted') {
    show.value = false
  }
  deferredPrompt = null
}

function dismiss() {
  show.value = false
  localStorage.setItem(DISMISS_KEY, 'true')
}

onMounted(() => {
  window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt)
  window.addEventListener('appinstalled', () => {
    show.value = false
    deferredPrompt = null
  })
})

onUnmounted(() => {
  window.removeEventListener('beforeinstallprompt', handleBeforeInstallPrompt)
})
</script>

<template>
  <Transition name="install-banner">
    <div v-if="show" class="install-banner">
      <span class="install-text">Manty Messenger를 홈 화면에 추가하세요</span>
      <div class="install-actions">
        <button class="install-btn" @click="install">설치</button>
        <button class="dismiss-btn" @click="dismiss">닫기</button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.install-banner {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: #007bff;
  color: white;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  z-index: 9999;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.15);
}

.install-text {
  font-size: 14px;
  font-weight: 500;
}

.install-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.install-btn {
  padding: 6px 16px;
  background: white;
  color: #007bff;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.dismiss-btn {
  padding: 6px 12px;
  background: transparent;
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.5);
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
}

.install-banner-enter-active,
.install-banner-leave-active {
  transition: transform 0.3s ease;
}
.install-banner-enter-from,
.install-banner-leave-to {
  transform: translateY(100%);
}
</style>
