<script setup>
import { ref } from 'vue'
import { uploadFile, getFileType } from '../services/api'

const emit = defineEmits(['send', 'sendFile', 'typing'])

const message = ref('')
const typingTimeout = ref(null)
const fileInput = ref(null)
const selectedFile = ref(null)
const filePreview = ref(null)
const uploadProgress = ref(0)
const isUploading = ref(false)

const MAX_FILE_SIZE = 100 * 1024 * 1024 // 100MB

const handleInput = () => {
  emit('typing', true)

  if (typingTimeout.value) {
    clearTimeout(typingTimeout.value)
  }

  typingTimeout.value = setTimeout(() => {
    emit('typing', false)
  }, 1500)
}

const handleSend = async () => {
  if (isUploading.value) return

  // Send file message
  if (selectedFile.value) {
    await sendFileMessage()
    return
  }

  // Send text message
  if (message.value.trim()) {
    emit('send', message.value)
    message.value = ''
    emit('typing', false)

    if (typingTimeout.value) {
      clearTimeout(typingTimeout.value)
    }
  }
}

const sendFileMessage = async () => {
  if (!selectedFile.value || isUploading.value) return

  isUploading.value = true
  uploadProgress.value = 0

  try {
    const result = await uploadFile(selectedFile.value, (progress) => {
      uploadProgress.value = progress
    })

    const messageType = getFileType(selectedFile.value.type)
    const content = message.value.trim() || selectedFile.value.name

    emit('sendFile', {
      content,
      messageType,
      fileUrl: result.url,
      thumbnailUrl: result.thumbnail_url
    })

    // Reset state
    clearFile()
    message.value = ''
  } catch (error) {
    console.error('File upload failed:', error)
    alert('파일 업로드에 실패했습니다.')
  } finally {
    isUploading.value = false
    uploadProgress.value = 0
  }
}

const handleKeydown = (e) => {
  if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
    e.preventDefault()
    handleSend()
  }
}

const openFilePicker = () => {
  fileInput.value?.click()
}

const handleFileSelect = (e) => {
  const file = e.target.files?.[0]
  if (!file) return

  if (file.size > MAX_FILE_SIZE) {
    alert('파일 크기는 100MB를 초과할 수 없습니다.')
    return
  }

  selectedFile.value = file

  // Create preview for images
  if (file.type.startsWith('image/')) {
    const reader = new FileReader()
    reader.onload = (e) => {
      filePreview.value = e.target.result
    }
    reader.readAsDataURL(file)
  } else {
    filePreview.value = null
  }

  // Reset input
  e.target.value = ''
}

const clearFile = () => {
  selectedFile.value = null
  filePreview.value = null
}

const formatFileSize = (bytes) => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

const getFileIcon = (mimeType) => {
  if (mimeType?.startsWith('image/')) return 'image'
  if (mimeType?.includes('pdf')) return 'pdf'
  if (mimeType?.includes('word') || mimeType?.includes('document')) return 'doc'
  if (mimeType?.includes('sheet') || mimeType?.includes('excel')) return 'xls'
  return 'file'
}
</script>

<template>
  <div class="message-input-container">
    <!-- File Preview -->
    <div v-if="selectedFile" class="file-preview">
      <div class="preview-content">
        <img v-if="filePreview" :src="filePreview" class="preview-image" />
        <div v-else class="preview-file">
          <div class="file-icon" :class="getFileIcon(selectedFile.type)">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
              <path d="M14 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V8l-6-6zM6 20V4h7v5h5v11H6z"/>
            </svg>
          </div>
          <div class="file-info">
            <span class="file-name">{{ selectedFile.name }}</span>
            <span class="file-size">{{ formatFileSize(selectedFile.size) }}</span>
          </div>
        </div>
      </div>
      <button @click="clearFile" class="clear-button" :disabled="isUploading">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
          <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
        </svg>
      </button>
      <!-- Upload Progress -->
      <div v-if="isUploading" class="upload-progress">
        <div class="progress-bar" :style="{ width: uploadProgress + '%' }"></div>
        <span class="progress-text">{{ uploadProgress }}%</span>
      </div>
    </div>

    <div class="input-wrapper">
      <!-- File Attach Button -->
      <button @click="openFilePicker" class="attach-button" :disabled="isUploading">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M16.5 6v11.5c0 2.21-1.79 4-4 4s-4-1.79-4-4V5c0-1.38 1.12-2.5 2.5-2.5s2.5 1.12 2.5 2.5v10.5c0 .55-.45 1-1 1s-1-.45-1-1V6H10v9.5c0 1.38 1.12 2.5 2.5 2.5s2.5-1.12 2.5-2.5V5c0-2.21-1.79-4-4-4S7 2.79 7 5v12.5c0 3.04 2.46 5.5 5.5 5.5s5.5-2.46 5.5-5.5V6h-1.5z"/>
        </svg>
      </button>
      <input
        type="file"
        ref="fileInput"
        @change="handleFileSelect"
        class="file-input"
        accept="image/*,.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.zip,.rar"
      />

      <textarea
        v-model="message"
        @input="handleInput"
        @keydown="handleKeydown"
        :placeholder="selectedFile ? '파일 설명을 입력하세요 (선택사항)' : '메시지를 입력하세요...'"
        rows="1"
        class="message-textarea"
        :disabled="isUploading"
      ></textarea>
      <button
        @click="handleSend"
        :disabled="(!message.trim() && !selectedFile) || isUploading"
        class="send-button"
      >
        <svg v-if="!isUploading" width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
        </svg>
        <svg v-else class="spinner" width="20" height="20" viewBox="0 0 24 24">
          <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" fill="none" stroke-dasharray="31.4 31.4" />
        </svg>
      </button>
    </div>
  </div>
</template>

<style scoped>
.message-input-container {
  padding: 16px 20px;
  background: white;
  border-top: 1px solid #e0e0e0;
}

.file-preview {
  position: relative;
  margin-bottom: 12px;
  padding: 12px;
  background: #f5f5f5;
  border-radius: 12px;
}

.preview-content {
  display: flex;
  align-items: center;
}

.preview-image {
  max-width: 200px;
  max-height: 150px;
  border-radius: 8px;
  object-fit: cover;
}

.preview-file {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  background: #e0e0e0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
}

.file-icon.pdf {
  background: #ffebee;
  color: #c62828;
}

.file-icon.doc {
  background: #e3f2fd;
  color: #1565c0;
}

.file-icon.xls {
  background: #e8f5e9;
  color: #2e7d32;
}

.file-icon.image {
  background: #fff3e0;
  color: #ef6c00;
}

.file-info {
  display: flex;
  flex-direction: column;
}

.file-name {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  font-size: 12px;
  color: #999;
}

.clear-button {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: none;
  background: rgba(0, 0, 0, 0.5);
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.clear-button:hover {
  background: rgba(0, 0, 0, 0.7);
}

.clear-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.upload-progress {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: #e0e0e0;
  border-radius: 0 0 12px 12px;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  background: #007bff;
  transition: width 0.2s ease;
}

.progress-text {
  position: absolute;
  right: 12px;
  bottom: 8px;
  font-size: 11px;
  color: #666;
}

.input-wrapper {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  background: #f5f5f5;
  border-radius: 24px;
  padding: 8px 8px 8px 12px;
}

.attach-button {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: transparent;
  color: #666;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: background-color 0.2s, color 0.2s;
}

.attach-button:hover:not(:disabled) {
  background: #e0e0e0;
  color: #333;
}

.attach-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.file-input {
  display: none;
}

.message-textarea {
  flex: 1;
  border: none;
  background: transparent;
  resize: none;
  font-size: 15px;
  line-height: 1.5;
  padding: 8px 0;
  max-height: 120px;
  font-family: inherit;
}

.message-textarea:focus {
  outline: none;
}

.message-textarea::placeholder {
  color: #999;
}

.message-textarea:disabled {
  opacity: 0.7;
}

.send-button {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: none;
  background: #007bff;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s, transform 0.1s;
  flex-shrink: 0;
}

.send-button:hover:not(:disabled) {
  background: #0056b3;
}

.send-button:active:not(:disabled) {
  transform: scale(0.95);
}

.send-button:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
