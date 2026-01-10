<script setup>
import { ref, watch } from 'vue'
import { useChatStore } from '../stores/chat'
import api from '../services/api'

const props = defineProps({
  roomId: {
    type: Number,
    required: true
  }
})

const emit = defineEmits(['close', 'invited'])

const chatStore = useChatStore()

const searchQuery = ref('')
const searchResults = ref([])
const searching = ref(false)
const inviting = ref(false)
const error = ref('')

let searchTimeout = null

watch(searchQuery, (value) => {
  if (searchTimeout) clearTimeout(searchTimeout)

  if (value.length < 2) {
    searchResults.value = []
    return
  }

  searchTimeout = setTimeout(async () => {
    searching.value = true
    try {
      searchResults.value = await chatStore.searchUsers(value)
    } catch (e) {
      console.error('Search failed', e)
    } finally {
      searching.value = false
    }
  }, 300)
})

const inviteUser = async (user) => {
  inviting.value = true
  error.value = ''

  try {
    await api.post(`/rooms/${props.roomId}/members`, {
      user_id: user.id
    })
    emit('invited', user)
  } catch (e) {
    error.value = e.response?.data?.error || '초대에 실패했습니다'
  } finally {
    inviting.value = false
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

const getInitial = (name) => {
  return name?.charAt(0).toUpperCase() || '?'
}
</script>

<template>
  <div class="modal-backdrop" @click="handleBackdropClick">
    <div class="modal-content card">
      <div class="modal-header">
        <h2>멤버 초대</h2>
        <button @click="handleClose" class="close-button">&times;</button>
      </div>

      <div class="search-section">
        <input
          v-model="searchQuery"
          type="text"
          class="input"
          placeholder="사용자 이름 또는 이메일로 검색..."
          autofocus
        />
      </div>

      <div class="results-section">
        <div v-if="searching" class="loading">검색 중...</div>

        <div v-else-if="searchQuery.length < 2" class="hint">
          2글자 이상 입력하세요
        </div>

        <div v-else-if="searchResults.length === 0" class="no-results">
          검색 결과가 없습니다
        </div>

        <div v-else class="user-list">
          <div
            v-for="user in searchResults"
            :key="user.id"
            class="user-item"
          >
            <div class="user-avatar">
              {{ getInitial(user.username) }}
            </div>
            <div class="user-info">
              <div class="user-name">{{ user.username }}</div>
              <div class="user-email">{{ user.email }}</div>
            </div>
            <button
              @click="inviteUser(user)"
              class="btn btn-primary invite-btn"
              :disabled="inviting"
            >
              초대
            </button>
          </div>
        </div>
      </div>

      <p v-if="error" class="error-message">{{ error }}</p>
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
  max-height: 80vh;
  margin: 20px;
  display: flex;
  flex-direction: column;
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
  margin-bottom: 20px;
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

.search-section {
  margin-bottom: 16px;
}

.results-section {
  flex: 1;
  overflow-y: auto;
  min-height: 200px;
  max-height: 400px;
}

.loading,
.hint,
.no-results {
  text-align: center;
  color: #888;
  padding: 40px 20px;
}

.user-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.user-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  background: #f8f9fa;
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 16px;
  flex-shrink: 0;
}

.user-info {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-weight: 600;
  color: #333;
}

.user-email {
  font-size: 13px;
  color: #888;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.invite-btn {
  padding: 8px 16px;
  font-size: 13px;
}

.error-message {
  color: #dc3545;
  text-align: center;
  font-size: 14px;
  margin-top: 12px;
}
</style>
