import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './assets/main.css'
import { initKeycloak } from './services/keycloak'
import { useAuthStore } from './stores/auth'

// Clean up URL fragment if it contains error from previous auth attempt
// This prevents the error fragment from being included in redirect_uri
if (window.location.hash && window.location.hash.includes('error=')) {
  window.history.replaceState({}, document.title, window.location.pathname + window.location.search)
}

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// Only the login page is a guest route, everything else requires auth
const isGuestRoute = window.location.pathname.includes('/login')

// Initialize Keycloak - require login for all protected routes
initKeycloak(!isGuestRoute).then(async (authenticated) => {
  app.mount('#app')

  if (authenticated) {
    const authStore = useAuthStore()
    await authStore.initAuth()
  }
})
