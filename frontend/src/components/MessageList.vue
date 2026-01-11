<script setup>
import { computed, ref, watch, nextTick, onMounted } from 'vue'
import { useChatStore } from '../stores/chat'
import { useAuthStore } from '../stores/auth'

const chatStore = useChatStore()
const authStore = useAuthStore()

const messagesContainer = ref(null)
const messages = computed(() => chatStore.currentMessages)
const currentUserId = computed(() => authStore.user?.id)

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

watch(messages, () => {
  scrollToBottom()
}, { deep: true })

onMounted(() => {
  scrollToBottom()
})

const formatTime = (dateString) => {
  const date = new Date(dateString)
  return date.toLocaleTimeString('ko-KR', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

const isMyMessage = (message) => {
  return message.sender?.id === currentUserId.value
}

const getInitial = (name) => {
  return name?.charAt(0).toUpperCase() || '?'
}
</script>

<template>
  <div class="messages-container" ref="messagesContainer">
    <div v-if="messages.length === 0" class="no-messages">
      <p>아직 메시지가 없습니다</p>
      <p class="hint">첫 메시지를 보내보세요!</p>
    </div>

    <div
      v-else
      v-for="message in messages"
      :key="message.id"
      class="message-wrapper"
      :class="{ 'my-message': isMyMessage(message) }"
    >
      <div v-if="!isMyMessage(message)" class="message-avatar">
        {{ getInitial(message.sender?.username) }}
      </div>

      <div class="message-content">
        <div v-if="!isMyMessage(message)" class="message-sender">
          {{ message.sender?.username }}
        </div>
        <div class="message-bubble">
          {{ message.content }}
        </div>
        <div class="message-time">
          <span v-if="message.unread_count > 0" class="unread-count">{{ message.unread_count }}</span>
          {{ formatTime(message.created_at) }}
          <span v-if="message.is_edited" class="edited">(수정됨)</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: #f8f9fa;
}

.no-messages {
  text-align: center;
  color: #666;
  padding: 40px 20px;
}

.hint {
  font-size: 13px;
  color: #999;
  margin-top: 8px;
}

.message-wrapper {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
  max-width: 70%;
}

.message-wrapper.my-message {
  flex-direction: row-reverse;
  margin-left: auto;
}

.message-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 14px;
  flex-shrink: 0;
}

.message-content {
  display: flex;
  flex-direction: column;
}

.my-message .message-content {
  align-items: flex-end;
}

.message-sender {
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
  margin-left: 4px;
}

.message-bubble {
  background: white;
  padding: 12px 16px;
  border-radius: 18px;
  border-top-left-radius: 4px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  word-break: break-word;
  line-height: 1.5;
}

.my-message .message-bubble {
  background: #007bff;
  color: white;
  border-radius: 18px;
  border-top-right-radius: 4px;
}

.message-time {
  font-size: 11px;
  color: #999;
  margin-top: 4px;
  margin-left: 4px;
}

.my-message .message-time {
  margin-right: 4px;
  margin-left: 0;
}

.edited {
  font-style: italic;
}

.unread-count {
  display: inline-block;
  background: #ffc107;
  color: #333;
  font-size: 10px;
  font-weight: bold;
  padding: 2px 6px;
  border-radius: 10px;
  margin-right: 4px;
}
</style>
