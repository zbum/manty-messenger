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

const isImageMessage = (message) => {
  return message.message_type === 'image' && message.file_url
}

const isFileMessage = (message) => {
  return message.message_type === 'file' && message.file_url
}

const getFileUrl = (url) => {
  if (!url) return ''
  // Handle relative URLs
  if (url.startsWith('/')) {
    return '/messenger' + url
  }
  return url
}

const getFileName = (message) => {
  // Try to get filename from content or URL
  if (message.content && message.content !== message.file_url) {
    return message.content
  }
  if (message.file_url) {
    const parts = message.file_url.split('/')
    return parts[parts.length - 1]
  }
  return 'file'
}

const getFileExtension = (url) => {
  if (!url) return ''
  const parts = url.split('.')
  return parts.length > 1 ? parts[parts.length - 1].toUpperCase() : ''
}

const downloadFile = (message) => {
  const url = getFileUrl(message.file_url)
  window.open(url, '_blank')
}

const openImage = (message) => {
  const url = getFileUrl(message.file_url)
  window.open(url, '_blank')
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

        <!-- Image Message -->
        <div v-if="isImageMessage(message)" class="message-bubble image-bubble" @click="openImage(message)">
          <img :src="getFileUrl(message.file_url)" :alt="message.content" class="message-image" />
          <div v-if="message.content && message.content !== getFileName(message)" class="image-caption">
            {{ message.content }}
          </div>
        </div>

        <!-- File Message -->
        <div v-else-if="isFileMessage(message)" class="message-bubble file-bubble" @click="downloadFile(message)">
          <div class="file-content">
            <div class="file-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
                <path d="M14 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V8l-6-6zM6 20V4h7v5h5v11H6z"/>
              </svg>
              <span class="file-ext">{{ getFileExtension(message.file_url) }}</span>
            </div>
            <div class="file-info">
              <span class="file-name">{{ getFileName(message) }}</span>
              <span class="file-action">클릭하여 다운로드</span>
            </div>
          </div>
        </div>

        <!-- Text Message -->
        <div v-else class="message-bubble">
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

/* Image Message Styles */
.image-bubble {
  padding: 4px;
  cursor: pointer;
  max-width: 300px;
}

.my-message .image-bubble {
  background: #007bff;
}

.message-image {
  max-width: 100%;
  max-height: 300px;
  border-radius: 14px;
  display: block;
  object-fit: cover;
}

.image-caption {
  padding: 8px 12px 4px;
  font-size: 14px;
}

/* File Message Styles */
.file-bubble {
  padding: 12px;
  cursor: pointer;
  min-width: 200px;
}

.file-bubble:hover {
  background: #f0f0f0;
}

.my-message .file-bubble:hover {
  background: #0056b3;
}

.file-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  background: #e8f5e9;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #2e7d32;
  flex-shrink: 0;
}

.my-message .file-icon {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

.file-ext {
  font-size: 8px;
  font-weight: bold;
  margin-top: -4px;
}

.file-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.file-name {
  font-size: 14px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-action {
  font-size: 12px;
  opacity: 0.7;
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
