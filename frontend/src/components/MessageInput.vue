<script setup>
import { ref } from 'vue'

const emit = defineEmits(['send', 'typing'])

const message = ref('')
const typingTimeout = ref(null)

const handleInput = () => {
  emit('typing', true)

  if (typingTimeout.value) {
    clearTimeout(typingTimeout.value)
  }

  typingTimeout.value = setTimeout(() => {
    emit('typing', false)
  }, 1500)
}

const handleSend = () => {
  if (message.value.trim()) {
    emit('send', message.value)
    message.value = ''
    emit('typing', false)

    if (typingTimeout.value) {
      clearTimeout(typingTimeout.value)
    }
  }
}

const handleKeydown = (e) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}
</script>

<template>
  <div class="message-input-container">
    <div class="input-wrapper">
      <textarea
        v-model="message"
        @input="handleInput"
        @keydown="handleKeydown"
        placeholder="메시지를 입력하세요..."
        rows="1"
        class="message-textarea"
      ></textarea>
      <button
        @click="handleSend"
        :disabled="!message.trim()"
        class="send-button"
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
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

.input-wrapper {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  background: #f5f5f5;
  border-radius: 24px;
  padding: 8px 8px 8px 16px;
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
</style>
