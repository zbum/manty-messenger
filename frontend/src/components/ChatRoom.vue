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

const handleSendFile = ({ content, messageType, fileUrl }) => {
  chatStore.sendFileMessage(content, messageType, fileUrl)
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
    <!-- Room Header -->
    <header class="room-header">
      <div class="room-title">
        <h2>{{ currentRoom?.name }}</h2>
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
</style>
