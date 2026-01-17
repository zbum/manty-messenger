<script setup>
import { ref } from 'vue'
import { stickerPacks } from '../data/stickers'

const emit = defineEmits(['select', 'close'])

const activePackId = ref(stickerPacks[0]?.id || '')

const activePack = () => {
  return stickerPacks.find(pack => pack.id === activePackId.value)
}

const selectSticker = (sticker) => {
  emit('select', sticker)
}

const selectPack = (packId) => {
  activePackId.value = packId
}
</script>

<template>
  <div class="sticker-picker">
    <!-- Sticker Grid -->
    <div class="sticker-grid">
      <button
        v-for="sticker in activePack()?.stickers"
        :key="sticker.id"
        class="sticker-item"
        @click="selectSticker(sticker)"
        :title="sticker.name"
      >
        <span class="sticker-emoji">{{ sticker.emoji }}</span>
      </button>
    </div>

    <!-- Pack Tabs -->
    <div class="pack-tabs">
      <button
        v-for="pack in stickerPacks"
        :key="pack.id"
        class="pack-tab"
        :class="{ active: pack.id === activePackId }"
        @click="selectPack(pack.id)"
        :title="pack.name"
      >
        {{ pack.icon }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.sticker-picker {
  position: absolute;
  bottom: 100%;
  left: 0;
  right: 0;
  background: white;
  border-top: 1px solid #e0e0e0;
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.1);
  z-index: 100;
}

.sticker-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 2px;
  padding: 8px;
  max-height: 240px;
  overflow-y: auto;
}

.sticker-item {
  width: 100%;
  aspect-ratio: 1;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s, transform 0.1s;
}

.sticker-item:hover {
  background: #f0f0f0;
}

.sticker-item:active {
  transform: scale(0.9);
}

.sticker-emoji {
  font-size: 36px;
  line-height: 1;
}

.pack-tabs {
  display: flex;
  border-top: 1px solid #e0e0e0;
  background: #fafafa;
  overflow-x: auto;
  scrollbar-width: none;
}

.pack-tabs::-webkit-scrollbar {
  display: none;
}

.pack-tab {
  flex: 0 0 auto;
  width: 48px;
  height: 44px;
  border: none;
  background: transparent;
  font-size: 20px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
  position: relative;
}

.pack-tab:hover {
  background: #e8e8e8;
}

.pack-tab.active {
  background: white;
}

.pack-tab.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 24px;
  height: 3px;
  background: #007bff;
  border-radius: 3px 3px 0 0;
}

/* Mobile Responsive */
@media (max-width: 768px) {
  .sticker-grid {
    grid-template-columns: repeat(6, 1fr);
    gap: 2px;
    padding: 6px;
    max-height: 200px;
  }

  .sticker-emoji {
    font-size: 32px;
  }

  .pack-tab {
    width: 44px;
    height: 40px;
    font-size: 18px;
  }
}

@media (max-width: 375px) {
  .sticker-grid {
    grid-template-columns: repeat(5, 1fr);
    max-height: 180px;
  }

  .sticker-emoji {
    font-size: 28px;
  }
}
</style>
