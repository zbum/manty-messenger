<script setup>
import { onMounted } from 'vue'
import { login, isAuthenticated } from '../services/keycloak'
import { useRouter } from 'vue-router'

const router = useRouter()

onMounted(() => {
  if (isAuthenticated()) {
    router.push('/chat')
  } else {
    login()
  }
})
</script>

<template>
  <div class="auth-container">
    <div class="auth-card card">
      <h1 class="auth-title">Keycloak 로그인으로 이동 중...</h1>
      <p class="loading-text">잠시만 기다려주세요.</p>
    </div>
  </div>
</template>

<style scoped>
.auth-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.auth-card {
  width: 100%;
  max-width: 400px;
  margin: 20px;
  text-align: center;
}

.auth-title {
  margin-bottom: 20px;
  color: #333;
}

.loading-text {
  color: #666;
}
</style>
