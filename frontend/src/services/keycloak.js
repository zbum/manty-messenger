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
        checkLoginIframe: false
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
  setInterval(() => {
    keycloak.updateToken(70)
      .catch(() => {
        console.warn('Failed to refresh token')
      })
  }, 60000)
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
