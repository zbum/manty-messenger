<script setup>
import { ref, watch, onMounted, computed } from 'vue'
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
const currentMembers = ref([])
const loadingMembers = ref(false)

// 현재 멤버 ID 목록 (검색 결과 필터링용)
const memberUserIds = computed(() => {
  return new Set(currentMembers.value.map(m => m.user.id))
})

let searchTimeout = null

// 모달이 열릴 때 현재 멤버 목록 가져오기
onMounted(async () => {
  loadingMembers.value = true
  try {
    const response = await api.get(`/rooms/${props.roomId}/members`)
    currentMembers.value = response.data || []
  } catch (e) {
    console.error('Failed to load members', e)
  } finally {
    loadingMembers.value = false
  }
})

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

      <!-- 현재 멤버 목록 섹션 -->
      <div class="members-section">
        <div class="section-header">
          <span class="section-title">현재 멤버</span>
          <span class="member-count">{{ currentMembers.length }}명</span>
        </div>
        <div v-if="loadingMembers" class="members-loading">불러오는 중...</div>
        <div v-else-if="currentMembers.length === 0" class="no-members">멤버가 없습니다</div>
        <div v-else class="members-list">
          <div
            v-for="member in currentMembers"
            :key="member.user.id"
            class="member-chip"
          >
            <div class="member-avatar">{{ getInitial(member.user.username) }}</div>
            <span class="member-name">{{ member.user.username }}</span>
            <span v-if="member.role === 'owner'" class="role-badge owner">방장</span>
            <span v-else-if="member.role === 'admin'" class="role-badge admin">관리자</span>
          </div>
        </div>
      </div>

      <div class="divider"></div>

      <div class="search-section">
        <div class="section-header">
          <span class="section-title">새 멤버 초대</span>
        </div>
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
            <span v-if="memberUserIds.has(user.id)" class="already-member">
              참여중
            </span>
            <button
              v-else
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

/* 현재 멤버 섹션 스타일 */
.members-section {
  margin-bottom: 8px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #555;
}

.member-count {
  font-size: 13px;
  color: #888;
}

.members-loading,
.no-members {
  font-size: 13px;
  color: #888;
  padding: 8px 0;
}

.members-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  max-height: 120px;
  overflow-y: auto;
}

.member-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background: #f0f2f5;
  border-radius: 20px;
  font-size: 13px;
}

.member-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 11px;
  flex-shrink: 0;
}

.member-name {
  color: #333;
  font-weight: 500;
}

.role-badge {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 10px;
  font-weight: 500;
}

.role-badge.owner {
  background: #fff3cd;
  color: #856404;
}

.role-badge.admin {
  background: #d1ecf1;
  color: #0c5460;
}

.divider {
  height: 1px;
  background: #e9ecef;
  margin: 16px 0;
}

.already-member {
  font-size: 13px;
  color: #28a745;
  font-weight: 500;
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

/* Mobile Responsive Styles */
@media (max-width: 768px) {
  .modal-backdrop {
    align-items: flex-end;
  }

  .modal-content {
    max-width: 100%;
    margin: 0;
    border-radius: 16px 16px 0 0;
    max-height: 85vh;
    padding-bottom: max(20px, env(safe-area-inset-bottom));
    animation: slideUpMobile 0.3s ease-out;
  }

  @keyframes slideUpMobile {
    from {
      opacity: 0;
      transform: translateY(100%);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .modal-header {
    margin-bottom: 16px;
  }

  .modal-header h2 {
    font-size: 18px;
  }

  .close-button {
    font-size: 24px;
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .search-section {
    margin-bottom: 12px;
  }

  .results-section {
    min-height: 150px;
    max-height: 50vh;
  }

  .user-item {
    padding: 10px;
    gap: 10px;
  }

  .user-avatar {
    width: 36px;
    height: 36px;
    font-size: 14px;
  }

  .user-name {
    font-size: 14px;
  }

  .user-email {
    font-size: 12px;
  }

  .invite-btn {
    padding: 8px 12px;
    font-size: 12px;
  }

  .loading,
  .hint,
  .no-results {
    padding: 30px 20px;
  }

  .members-section {
    margin-bottom: 4px;
  }

  .section-header {
    margin-bottom: 8px;
  }

  .section-title {
    font-size: 13px;
  }

  .members-list {
    max-height: 100px;
    gap: 6px;
  }

  .member-chip {
    padding: 4px 8px;
    font-size: 12px;
    gap: 4px;
  }

  .member-avatar {
    width: 20px;
    height: 20px;
    font-size: 10px;
  }

  .divider {
    margin: 12px 0;
  }

  .already-member {
    font-size: 12px;
  }
}
</style>
