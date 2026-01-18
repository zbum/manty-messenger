<script setup>
import { computed, ref, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useChatStore } from '../stores/chat'
import { useAuthStore } from '../stores/auth'
import { getStickerById } from '../data/stickers'

const chatStore = useChatStore()
const authStore = useAuthStore()

const messagesContainer = ref(null)
const messages = computed(() => chatStore.currentMessages)
const currentUserId = computed(() => authStore.user?.id)
const currentRoom = computed(() => chatStore.currentRoom)
const loadingMore = computed(() => chatStore.loadingMore)
const hasMore = computed(() => currentRoom.value ? chatStore.hasMore[currentRoom.value.id] : false)

// 스크롤 위치 추적
let isLoadingOlder = false
let previousScrollHeight = 0

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

// 이미지 로드 완료 시 스크롤 (이미지 높이 반영 후)
const onImageLoad = () => {
  scrollToBottom()
}

// 스크롤 이벤트 핸들러 (무한 스크롤)
const handleScroll = async () => {
  if (!messagesContainer.value || !currentRoom.value) return
  if (loadingMore.value || !hasMore.value) return

  const { scrollTop } = messagesContainer.value

  // 상단에서 100px 이내로 스크롤하면 이전 메시지 로드
  if (scrollTop < 100) {
    isLoadingOlder = true
    previousScrollHeight = messagesContainer.value.scrollHeight

    const loaded = await chatStore.loadMoreMessages(currentRoom.value.id)

    if (loaded) {
      // 스크롤 위치 유지 (새로 추가된 메시지 높이만큼 보정)
      await nextTick()
      const newScrollHeight = messagesContainer.value.scrollHeight
      messagesContainer.value.scrollTop = newScrollHeight - previousScrollHeight
    }

    isLoadingOlder = false
  }
}

// 메시지 변경 감지
watch(messages, (newMessages, oldMessages) => {
  // 이전 메시지 로딩 중이 아닐 때만 하단으로 스크롤
  if (!isLoadingOlder) {
    scrollToBottom()
  }
}, { deep: true })

// 방 변경 시 하단으로 스크롤
watch(currentRoom, () => {
  scrollToBottom()
})

onMounted(() => {
  scrollToBottom()
  if (messagesContainer.value) {
    messagesContainer.value.addEventListener('scroll', handleScroll)
  }
})

onUnmounted(() => {
  if (messagesContainer.value) {
    messagesContainer.value.removeEventListener('scroll', handleScroll)
  }
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

const isStickerMessage = (message) => {
  return message.message_type === 'sticker'
}

const getStickerEmoji = (message) => {
  const sticker = getStickerById(message.content)
  return sticker?.emoji || message.content
}

const getFileUrl = (url) => {
  if (!url) return ''
  // Handle relative URLs
  if (url.startsWith('/')) {
    return '/messenger' + url
  }
  return url
}

const getThumbnailUrl = (message) => {
  // Use thumbnail_url if available, otherwise fall back to file_url
  if (message.thumbnail_url) {
    return getFileUrl(message.thumbnail_url)
  }
  return getFileUrl(message.file_url)
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
  const link = document.createElement('a')
  link.href = url
  link.download = getFileName(message)
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

const openImage = (message) => {
  const url = getFileUrl(message.file_url)
  window.open(url, '_blank')
}
</script>

<template>
  <div class="messages-container" ref="messagesContainer">
    <!-- 이전 메시지 로딩 표시 -->
    <div v-if="loadingMore" class="loading-more">
      <div class="loading-spinner"></div>
      <span>이전 메시지 불러오는 중...</span>
    </div>

    <!-- 더 이상 메시지가 없을 때 -->
    <div v-else-if="messages.length > 0 && !hasMore" class="no-more-messages">
      대화의 시작입니다
    </div>

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

        <!-- Sticker Message -->
        <div v-if="isStickerMessage(message)" class="sticker-bubble">
          <span class="sticker-emoji">{{ getStickerEmoji(message) }}</span>
        </div>

        <!-- Image Message -->
        <div v-else-if="isImageMessage(message)" class="message-bubble image-bubble" @click="openImage(message)">
          <img :src="getThumbnailUrl(message)" :alt="message.content" class="message-image" @load="onImageLoad" />
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

/* 이전 메시지 로딩 스타일 */
.loading-more {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 16px;
  color: #666;
  font-size: 13px;
}

.loading-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid #e0e0e0;
  border-top-color: #007bff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.no-more-messages {
  text-align: center;
  color: #999;
  font-size: 12px;
  padding: 16px;
  margin-bottom: 8px;
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

/* Sticker Message Styles */
.sticker-bubble {
  background: transparent;
  padding: 0;
}

.sticker-emoji {
  font-size: 80px;
  line-height: 1;
  display: block;
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

/* Mobile Responsive Styles */
@media (max-width: 768px) {
  .messages-container {
    padding: 12px;
  }

  .message-wrapper {
    max-width: 85%;
    gap: 8px;
    margin-bottom: 12px;
  }

  .message-avatar {
    width: 32px;
    height: 32px;
    font-size: 12px;
  }

  .message-bubble {
    padding: 10px 14px;
    border-radius: 16px;
    font-size: 15px;
  }

  .image-bubble {
    max-width: 240px;
  }

  .message-image {
    max-height: 200px;
  }

  .sticker-emoji {
    font-size: 64px;
  }

  .file-bubble {
    padding: 10px;
    min-width: 180px;
  }

  .file-icon {
    width: 40px;
    height: 40px;
  }

  .file-icon svg {
    width: 20px;
    height: 20px;
  }

  .file-name {
    font-size: 13px;
  }

  .file-action {
    font-size: 11px;
  }

  .message-sender {
    font-size: 11px;
  }

  .message-time {
    font-size: 10px;
  }
}

@media (max-width: 375px) {
  .message-wrapper {
    max-width: 90%;
  }

  .image-bubble {
    max-width: 200px;
  }

  .file-bubble {
    min-width: 160px;
  }
}
</style>
