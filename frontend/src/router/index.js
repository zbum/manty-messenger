import { createRouter, createWebHistory } from 'vue-router'
import { isAuthenticated } from '../services/keycloak'

const routes = [
  {
    path: '/',
    redirect: '/chat'
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/LoginView.vue'),
    meta: { guest: true }
  },
  {
    path: '/chat',
    name: 'chat',
    component: () => import('../views/ChatView.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory('/messenger/'),
  routes
})

router.beforeEach(async (to, from, next) => {
  // Guest routes redirect to chat if authenticated
  if (to.meta.guest && isAuthenticated()) {
    next('/chat')
    return
  }

  next()
})

export default router
