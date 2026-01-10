<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useChatStore } from '../stores/chat'

const router = useRouter()
const authStore = useAuthStore()
const chatStore = useChatStore()

const email = ref('')
const password = ref('')

const handleSubmit = async () => {
  const success = await authStore.login(email.value, password.value)
  if (success) {
    chatStore.initWebSocketListeners()
    router.push('/chat')
  }
}
</script>

<template>
  <div class="auth-container">
    <div class="auth-card card">
      <h1 class="auth-title">로그인</h1>

      <form @submit.prevent="handleSubmit" class="auth-form">
        <div class="form-group">
          <label for="email">이메일</label>
          <input
            id="email"
            v-model="email"
            type="email"
            class="input"
            placeholder="이메일을 입력하세요"
            required
          />
        </div>

        <div class="form-group">
          <label for="password">비밀번호</label>
          <input
            id="password"
            v-model="password"
            type="password"
            class="input"
            placeholder="비밀번호를 입력하세요"
            required
          />
        </div>

        <p v-if="authStore.error" class="error-message">
          {{ authStore.error }}
        </p>

        <button
          type="submit"
          class="btn btn-primary submit-btn"
          :disabled="authStore.loading"
        >
          {{ authStore.loading ? '로그인 중...' : '로그인' }}
        </button>
      </form>

      <p class="auth-link">
        계정이 없으신가요?
        <router-link to="/register">회원가입</router-link>
      </p>
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
}

.auth-title {
  text-align: center;
  margin-bottom: 30px;
  color: #333;
}

.auth-form {
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

.submit-btn {
  width: 100%;
  padding: 14px;
  font-size: 16px;
  margin-top: 10px;
}

.error-message {
  color: #dc3545;
  text-align: center;
  font-size: 14px;
}

.auth-link {
  text-align: center;
  margin-top: 20px;
  color: #666;
}

.auth-link a {
  color: #007bff;
  text-decoration: none;
  font-weight: 500;
}

.auth-link a:hover {
  text-decoration: underline;
}
</style>
