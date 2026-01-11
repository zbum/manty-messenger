<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
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

const hasCurrentRoom = computed(() => !!chatStore.currentRoom)

onMounted(async () => {
  await chatStore.fetchRooms()
  await chatStore.restoreLastRoom()
})

onUnmounted(() => {
  chatStore.leaveCurrentRoom()
})

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

const handleRoomCreated = (room) => {
  showCreateModal.value = false
  chatStore.joinRoom(room)
}
</script>

<template>
  <div class="chat-layout">
    <!-- Sidebar -->
    <aside class="sidebar">
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
    <main class="chat-main">
      <ChatRoom v-if="hasCurrentRoom" />
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
</style>
