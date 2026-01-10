<script setup>
import { ref } from 'vue'
import { useChatStore } from '../stores/chat'

const emit = defineEmits(['close', 'created'])

const chatStore = useChatStore()

const name = ref('')
const description = ref('')
const roomType = ref('group')
const loading = ref(false)
const error = ref('')

const handleSubmit = async () => {
  if (!name.value.trim()) {
    error.value = '채팅방 이름을 입력하세요'
    return
  }

  loading.value = true
  error.value = ''

  const room = await chatStore.createRoom(
    name.value.trim(),
    description.value.trim(),
    roomType.value
  )

  loading.value = false

  if (room) {
    emit('created', room)
  } else {
    error.value = chatStore.error || '채팅방 생성에 실패했습니다'
  }
}

const handleClose = () => {
  emit('close')
}

const handleBackdropClick = (e) => {
  if (e.target === e.currentTarget) {
    handleClose()
  }
}
</script>

<template>
  <div class="modal-backdrop" @click="handleBackdropClick">
    <div class="modal-content card">
      <div class="modal-header">
        <h2>새 채팅방 만들기</h2>
        <button @click="handleClose" class="close-button">&times;</button>
      </div>

      <form @submit.prevent="handleSubmit" class="modal-form">
        <div class="form-group">
          <label for="name">채팅방 이름 *</label>
          <input
            id="name"
            v-model="name"
            type="text"
            class="input"
            placeholder="채팅방 이름을 입력하세요"
            required
          />
        </div>

        <div class="form-group">
          <label for="description">설명</label>
          <textarea
            id="description"
            v-model="description"
            class="input textarea"
            placeholder="채팅방 설명을 입력하세요 (선택)"
            rows="3"
          ></textarea>
        </div>

        <div class="form-group">
          <label>채팅방 유형</label>
          <div class="radio-group">
            <label class="radio-label">
              <input
                type="radio"
                v-model="roomType"
                value="group"
              />
              <span>그룹 채팅</span>
            </label>
            <label class="radio-label">
              <input
                type="radio"
                v-model="roomType"
                value="private"
              />
              <span>비공개 채팅</span>
            </label>
          </div>
        </div>

        <p v-if="error" class="error-message">{{ error }}</p>

        <div class="modal-actions">
          <button type="button" @click="handleClose" class="btn btn-secondary">
            취소
          </button>
          <button type="submit" class="btn btn-primary" :disabled="loading">
            {{ loading ? '생성 중...' : '만들기' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  width: 100%;
  max-width: 450px;
  margin: 20px;
  animation: slideUp 0.2s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.modal-header h2 {
  font-size: 20px;
  color: #333;
  margin: 0;
}

.close-button {
  background: none;
  border: none;
  font-size: 28px;
  color: #999;
  cursor: pointer;
  padding: 0;
  line-height: 1;
}

.close-button:hover {
  color: #333;
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-weight: 500;
  color: #555;
}

.textarea {
  resize: vertical;
  min-height: 80px;
}

.radio-group {
  display: flex;
  gap: 20px;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-weight: normal;
}

.radio-label input {
  cursor: pointer;
}

.error-message {
  color: #dc3545;
  text-align: center;
  font-size: 14px;
  margin: 0;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 10px;
}
</style>
