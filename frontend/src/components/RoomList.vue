<script setup>
import { computed } from 'vue'
import { useChatStore } from '../stores/chat'

const chatStore = useChatStore()

const rooms = computed(() => chatStore.rooms)
const currentRoom = computed(() => chatStore.currentRoom)

const selectRoom = (room) => {
  chatStore.joinRoom(room)
}

const getInitial = (name) => {
  return name?.charAt(0).toUpperCase() || '?'
}
</script>

<template>
  <div class="room-list">
    <div v-if="chatStore.loading" class="loading">
      로딩 중...
    </div>

    <div v-else-if="rooms.length === 0" class="empty-state">
      <p>채팅방이 없습니다</p>
      <p class="hint">새 채팅방을 만들어보세요</p>
    </div>

    <div
      v-else
      v-for="room in rooms"
      :key="room.id"
      class="room-item"
      :class="{ active: currentRoom?.id === room.id }"
      @click="selectRoom(room)"
    >
      <div class="room-avatar">
        {{ getInitial(room.name) }}
      </div>
      <div class="room-info">
        <div class="room-name">{{ room.name }}</div>
        <div class="room-meta">
          <span class="member-count">{{ room.member_count || 1 }}명</span>
          <span class="room-type" :class="room.room_type">
            {{ room.room_type === 'private' ? '비공개' : '그룹' }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.room-list {
  flex: 1;
  overflow-y: auto;
}

.loading,
.empty-state {
  padding: 20px;
  text-align: center;
  color: #666;
}

.hint {
  font-size: 13px;
  color: #999;
  margin-top: 8px;
}

.room-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  cursor: pointer;
  transition: background-color 0.2s;
  border-bottom: 1px solid #f0f0f0;
}

.room-item:hover {
  background-color: #f8f9fa;
}

.room-item.active {
  background-color: #e3f2fd;
  border-left: 3px solid #007bff;
}

.room-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 18px;
  flex-shrink: 0;
}

.room-info {
  flex: 1;
  min-width: 0;
}

.room-name {
  font-weight: 600;
  color: #333;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.room-meta {
  display: flex;
  gap: 8px;
  margin-top: 4px;
  font-size: 12px;
  color: #888;
}

.room-type {
  padding: 2px 6px;
  border-radius: 4px;
  background: #f0f0f0;
}

.room-type.group {
  background: #e3f2fd;
  color: #1976d2;
}

.room-type.private {
  background: #fff3e0;
  color: #f57c00;
}
</style>
