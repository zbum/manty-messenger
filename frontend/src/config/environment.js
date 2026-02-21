export function isCapacitor() {
  return typeof window !== 'undefined' && window.Capacitor !== undefined
}

export function getPlatform() {
  if (!isCapacitor()) return 'web'
  return window.Capacitor.getPlatform()
}

export function isNative() {
  return isCapacitor() && window.Capacitor.isNativePlatform()
}

export function getApiBaseUrl() {
  if (isNative()) {
    return import.meta.env.VITE_API_BASE_URL || 'https://manty.co.kr/messenger/api/v1'
  }
  return '/messenger/api/v1'
}

export function getWsBaseUrl() {
  if (isNative()) {
    const apiBase = import.meta.env.VITE_API_BASE_URL || 'https://manty.co.kr/messenger'
    return apiBase.replace(/^http/, 'ws').replace(/\/api\/v1$/, '') + '/ws'
  }
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}/messenger/ws`
}

export function getKeycloakRedirectUri(path = '/chat') {
  if (isNative()) {
    return 'com.manty.messenger://auth/callback'
  }
  return window.location.origin + '/messenger' + path
}

export function getBaseUrl() {
  if (isNative()) {
    return '/'
  }
  return '/messenger/'
}
