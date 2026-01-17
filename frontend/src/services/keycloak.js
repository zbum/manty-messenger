import Keycloak from 'keycloak-js'

const keycloak = new Keycloak({
  url: 'https://keycloak.manty.co.kr',
  realm: 'manty',
  clientId: 'manty-messenger'
})

let initialized = false
let initPromise = null

export async function initKeycloak(requireLogin = false) {
  if (initialized) {
    return keycloak.authenticated
  }

  if (initPromise) {
    return initPromise
  }

  initPromise = new Promise(async (resolve) => {
    try {
      const authenticated = await keycloak.init({
        onLoad: requireLogin ? 'login-required' : 'check-sso',
        pkceMethod: 'S256',
        checkLoginIframe: false,
        silentCheckSsoRedirectUri: window.location.origin + '/messenger/silent-check-sso.html'
      })

      initialized = true

      if (authenticated) {
        console.log('Keycloak authenticated, token:', keycloak.token?.substring(0, 50) + '...')
        setupTokenRefresh()
      } else {
        console.log('Keycloak not authenticated')
      }

      resolve(authenticated)
    } catch (error) {
      console.error('Keycloak initialization failed:', error)
      initialized = true
      resolve(false)
    }
  })

  return initPromise
}

function setupTokenRefresh() {
  // 주기적 토큰 갱신 (1분마다)
  setInterval(() => {
    refreshTokenIfNeeded()
  }, 60000)

  // 페이지가 다시 활성화될 때 토큰 갱신
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') {
      console.log('Page visible, checking token...')
      refreshTokenIfNeeded()
    }
  })

  // 온라인 상태로 돌아올 때 토큰 갱신
  window.addEventListener('online', () => {
    console.log('Network online, checking token...')
    refreshTokenIfNeeded()
  })
}

async function refreshTokenIfNeeded() {
  if (!keycloak.authenticated) {
    return
  }

  try {
    // 토큰이 70초 이내에 만료되면 갱신
    const refreshed = await keycloak.updateToken(70)
    if (refreshed) {
      console.log('Token refreshed successfully')
    }
  } catch (error) {
    console.error('Failed to refresh token:', error)
    // 토큰 갱신 실패 시 재로그인 시도
    if (keycloak.isTokenExpired()) {
      console.warn('Token expired, redirecting to login...')
      keycloak.login({
        redirectUri: window.location.href
      })
    }
  }
}

export function login() {
  return keycloak.login({
    redirectUri: window.location.origin + '/messenger/chat'
  })
}

export function logout() {
  return keycloak.logout({
    redirectUri: window.location.origin + '/messenger/'
  })
}

export function getToken() {
  return keycloak.token
}

// 토큰을 가져오기 전에 갱신 확인 (async 버전)
export async function getValidToken() {
  if (!keycloak.authenticated) {
    return null
  }

  try {
    await keycloak.updateToken(30) // 30초 이내 만료면 갱신
    return keycloak.token
  } catch (error) {
    console.error('Failed to get valid token:', error)
    return null
  }
}

export function isAuthenticated() {
  return keycloak.authenticated === true
}

export function isInitialized() {
  return initialized
}

export function getUsername() {
  return keycloak.tokenParsed?.preferred_username
}

export function getEmail() {
  return keycloak.tokenParsed?.email
}

export default keycloak
