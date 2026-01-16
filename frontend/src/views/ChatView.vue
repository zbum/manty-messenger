<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useChatStore } from '../stores/chat'
import RoomList from '../components/RoomList.vue'
import ChatRoom from '../components/ChatRoom.vue'
import CreateRoomModal from '../components/CreateRoomModal.vue'

const router = useRouter()
const authStore = useAuthStore()
const chatStore = useChatStore()

const showCreateModal = ref(false)
const showSidebar = ref(true)

const hasCurrentRoom = computed(() => !!chatStore.currentRoom)

// Mobile detection
const isMobile = ref(window.innerWidth <= 768)

const handleResize = () => {
  isMobile.value = window.innerWidth <= 768
  // On desktop, always show sidebar
  if (!isMobile.value) {
    showSidebar.value = true
  }
}

onMounted(async () => {
  window.addEventListener('resize', handleResize)
  await chatStore.fetchRooms()
  await chatStore.restoreLastRoom()
  // If on mobile and has current room, hide sidebar
  if (isMobile.value && hasCurrentRoom.value) {
    showSidebar.value = false
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  chatStore.leaveCurrentRoom()
})

// Watch for room changes on mobile
watch(hasCurrentRoom, (hasRoom) => {
  if (isMobile.value && hasRoom) {
    showSidebar.value = false
  }
})

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

const handleRoomCreated = (room) => {
  showCreateModal.value = false
  chatStore.joinRoom(room)
}

const toggleSidebar = () => {
  showSidebar.value = !showSidebar.value
}

const handleBackToList = () => {
  showSidebar.value = true
}

// Expose for ChatRoom component
defineExpose({ handleBackToList, isMobile })
</script>

<template>
  <div class="chat-layout" :class="{ 'mobile': isMobile, 'sidebar-open': showSidebar }">
    <!-- Mobile Overlay -->
    <div
      v-if="isMobile && showSidebar && hasCurrentRoom"
      class="sidebar-overlay"
      @click="showSidebar = false"
    ></div>

    <!-- Sidebar -->
    <aside class="sidebar" :class="{ 'hidden': isMobile && !showSidebar }">
      <div class="sidebar-header">
        <div class="user-info">
          <div class="avatar">
            {{ authStore.user?.username?.charAt(0).toUpperCase() }}
          </div>
          <span class="username">{{ authStore.user?.username }}</span>
        </div>
        <button @click="handleLogout" class="btn btn-secondary logout-btn">
          로그아웃
        </button>
      </div>

      <div class="sidebar-actions">
        <button @click="showCreateModal = true" class="btn btn-primary create-room-btn">
          + 새 채팅방
        </button>
      </div>

      <RoomList />
    </aside>

    <!-- Main Chat Area -->
    <main class="chat-main" :class="{ 'hidden': isMobile && showSidebar && !hasCurrentRoom }">
      <ChatRoom v-if="hasCurrentRoom" @back="handleBackToList" :is-mobile="isMobile" />
      <div v-else class="no-room-selected">
        <div class="no-room-content">
          <h2>채팅방을 선택하세요</h2>
          <p>왼쪽에서 채팅방을 선택하거나 새 채팅방을 만드세요</p>
        </div>
      </div>
    </main>

    <!-- Create Room Modal -->
    <CreateRoomModal
      v-if="showCreateModal"
      @close="showCreateModal = false"
      @created="handleRoomCreated"
    />
  </div>
</template>

<style scoped>
.chat-layout {
  display: flex;
  height: 100vh;
  background-color: #f5f5f5;
}

.sidebar {
  width: 320px;
  background: white;
  border-right: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.username {
  font-weight: 600;
  color: #333;
}

.logout-btn {
  padding: 8px 12px;
  font-size: 12px;
}

.sidebar-actions {
  padding: 16px;
}

.create-room-btn {
  width: 100%;
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: white;
}

.no-room-selected {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f8f9fa;
}

.no-room-content {
  text-align: center;
  color: #666;
}

.no-room-content h2 {
  margin-bottom: 10px;
  color: #333;
}

/* Mobile Overlay */
.sidebar-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 99;
}

/* Mobile Responsive Styles */
@media (max-width: 768px) {
  .chat-layout {
    position: relative;
  }

  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    width: 100%;
    max-width: 320px;
    z-index: 100;
    transform: translateX(0);
    transition: transform 0.3s ease;
  }

  .sidebar.hidden {
    transform: translateX(-100%);
  }

  .chat-main {
    width: 100%;
    min-width: 0;
  }

  .chat-main.hidden {
    display: none;
  }

  .no-room-selected {
    display: none;
  }

  .chat-layout.mobile.sidebar-open .no-room-selected {
    display: flex;
  }

  .sidebar-header {
    padding: 12px 16px;
  }

  .logout-btn {
    padding: 6px 10px;
    font-size: 11px;
  }

  .sidebar-actions {
    padding: 12px 16px;
  }
}

/* Small mobile */
@media (max-width: 375px) {
  .sidebar {
    max-width: 100%;
  }

  .username {
    font-size: 14px;
    max-width: 120px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
