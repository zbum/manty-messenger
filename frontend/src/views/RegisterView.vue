<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useChatStore } from '../stores/chat'

const router = useRouter()
const authStore = useAuthStore()
const chatStore = useChatStore()

const email = ref('')
const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const localError = ref('')

const handleSubmit = async () => {
  localError.value = ''

  if (password.value !== confirmPassword.value) {
    localError.value = '비밀번호가 일치하지 않습니다'
    return
  }

  if (password.value.length < 6) {
    localError.value = '비밀번호는 최소 6자 이상이어야 합니다'
    return
  }

  const success = await authStore.register(email.value, username.value, password.value)
  if (success) {
    chatStore.initWebSocketListeners()
    router.push('/chat')
  }
}
</script>

<template>
  <div class="auth-container">
    <div class="auth-card card">
      <h1 class="auth-title">회원가입</h1>

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
          <label for="username">사용자 이름</label>
          <input
            id="username"
            v-model="username"
            type="text"
            class="input"
            placeholder="사용자 이름을 입력하세요"
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

        <div class="form-group">
          <label for="confirmPassword">비밀번호 확인</label>
          <input
            id="confirmPassword"
            v-model="confirmPassword"
            type="password"
            class="input"
            placeholder="비밀번호를 다시 입력하세요"
            required
          />
        </div>

        <p v-if="localError || authStore.error" class="error-message">
          {{ localError || authStore.error }}
        </p>

        <button
          type="submit"
          class="btn btn-primary submit-btn"
          :disabled="authStore.loading"
        >
          {{ authStore.loading ? '가입 중...' : '회원가입' }}
        </button>
      </form>

      <p class="auth-link">
        이미 계정이 있으신가요?
        <router-link to="/login">로그인</router-link>
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
