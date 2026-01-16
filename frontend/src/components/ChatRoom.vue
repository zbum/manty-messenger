<script setup>
import { ref, computed } from 'vue'
import { useChatStore } from '../stores/chat'
import { useAuthStore } from '../stores/auth'
import MessageList from './MessageList.vue'
import MessageInput from './MessageInput.vue'
import InviteMemberModal from './InviteMemberModal.vue'

const chatStore = useChatStore()
const authStore = useAuthStore()

const showInviteModal = ref(false)

const currentRoom = computed(() => chatStore.currentRoom)
const typingUsers = computed(() => chatStore.currentTypingUsers)

// 연결 상태
const isConnected = computed(() => chatStore.isConnected)
const isConnecting = computed(() => chatStore.isConnecting)
const isReconnecting = computed(() => chatStore.isReconnecting)
const isDisconnected = computed(() => chatStore.isDisconnected)
const offlineQueueCount = computed(() => chatStore.offlineQueueCount)

const connectionStatusText = computed(() => {
  if (isReconnecting.value) return '재연결 중...'
  if (isConnecting.value) return '연결 중...'
  if (!isConnected.value) return '연결 끊김'
  return ''
})

const connectionStatusClass = computed(() => {
  if (isReconnecting.value || isConnecting.value) return 'connecting'
  if (!isConnected.value) return 'disconnected'
  return 'connected'
})

// 재연결 버튼 표시 여부 (완전히 끊겼을 때만)
const showReconnectButton = computed(() => isDisconnected.value)

const handleReconnect = () => {
  chatStore.reconnect()
}

const typingText = computed(() => {
  if (typingUsers.value.length === 0) return ''
  if (typingUsers.value.length === 1) {
    return `${typingUsers.value[0].username}님이 입력 중...`
  }
  return `${typingUsers.value.length}명이 입력 중...`
})

const handleSendMessage = (content) => {
  chatStore.sendMessage(content)
}

const handleSendFile = ({ content, messageType, fileUrl, thumbnailUrl }) => {
  chatStore.sendFileMessage(content, messageType, fileUrl, thumbnailUrl)
}

const handleTyping = (isTyping) => {
  chatStore.setTyping(isTyping)
}

const handleInvited = (user) => {
  showInviteModal.value = false
  // member_count will be updated via WebSocket
}
</script>

<template>
  <div class="chat-room">
    <!-- Connection Status Banner -->
    <div v-if="connectionStatusText" :class="['connection-banner', connectionStatusClass]">
      <span class="connection-indicator"></span>
      <span>{{ connectionStatusText }}</span>
      <span v-if="offlineQueueCount > 0" class="offline-queue">
        (대기 중 메시지 {{ offlineQueueCount }}개)
      </span>
      <button v-if="showReconnectButton" @click="handleReconnect" class="reconnect-button">
        재연결
      </button>
    </div>

    <!-- Room Header -->
    <header class="room-header">
      <div class="room-title">
        <div class="room-name-row">
          <h2>{{ currentRoom?.name }}</h2>
          <span :class="['connection-dot', connectionStatusClass]" :title="connectionStatusText || '연결됨'"></span>
        </div>
        <span class="member-count">{{ currentRoom?.member_count || 1 }}명</span>
      </div>
      <button @click="showInviteModal = true" class="invite-button">
        + 초대
      </button>
    </header>

    <!-- Messages Area -->
    <MessageList />

    <!-- Typing Indicator -->
    <div v-if="typingText" class="typing-indicator">
      {{ typingText }}
    </div>

    <!-- Message Input -->
    <MessageInput
      @send="handleSendMessage"
      @sendFile="handleSendFile"
      @typing="handleTyping"
    />

    <!-- Invite Modal -->
    <InviteMemberModal
      v-if="showInviteModal"
      :room-id="currentRoom?.id"
      @close="showInviteModal = false"
      @invited="handleInvited"
    />
  </div>
</template>

<style scoped>
.chat-room {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.room-header {
  padding: 16px 20px;
  border-bottom: 1px solid #e0e0e0;
  background: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.room-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.invite-button {
  padding: 8px 16px;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background-color 0.2s;
}

.invite-button:hover {
  background: #0056b3;
}

.room-title h2 {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.member-count {
  font-size: 13px;
  color: #888;
  background: #f0f0f0;
  padding: 4px 8px;
  border-radius: 12px;
}

.typing-indicator {
  padding: 8px 20px;
  font-size: 13px;
  color: #666;
  font-style: italic;
  background: #fafafa;
}

/* Connection Status Styles */
.connection-banner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 500;
}

.connection-banner.disconnected {
  background: #fee2e2;
  color: #991b1b;
}

.connection-banner.connecting {
  background: #fef3c7;
  color: #92400e;
}

.connection-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.connection-banner.disconnected .connection-indicator {
  background: #dc2626;
}

.connection-banner.connecting .connection-indicator {
  background: #f59e0b;
  animation: pulse 1.5s infinite;
}

.offline-queue {
  font-size: 12px;
  opacity: 0.8;
}

.reconnect-button {
  margin-left: 8px;
  padding: 4px 12px;
  background: #991b1b;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.reconnect-button:hover {
  background: #7f1d1d;
}

.room-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.connection-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.connection-dot.connected {
  background: #22c55e;
}

.connection-dot.disconnected {
  background: #dc2626;
}

.connection-dot.connecting {
  background: #f59e0b;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
