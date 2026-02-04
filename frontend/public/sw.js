/**
 * Service Worker for PWA - Push Notifications & Offline Caching
 */

const CACHE_NAME = 'manty-v1'
const OFFLINE_URL = '/messenger/offline.html'

const PRECACHE_ASSETS = [
  '/messenger/',
  OFFLINE_URL
]

// 서비스 워커 설치 - 핵심 에셋 프리캐시
self.addEventListener('install', (event) => {
  console.log('Service Worker installed')
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(PRECACHE_ASSETS)
    })
  )
  self.skipWaiting()
})

// 서비스 워커 활성화 - 이전 캐시 정리
self.addEventListener('activate', (event) => {
  console.log('Service Worker activated')
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames
          .filter((name) => name !== CACHE_NAME)
          .map((name) => caches.delete(name))
      )
    }).then(() => clients.claim())
  )
})

// Fetch 이벤트 - 캐싱 전략 적용
self.addEventListener('fetch', (event) => {
  const { request } = event
  const url = new URL(request.url)

  // API 요청 - Network only (캐시하지 않음)
  if (url.pathname.startsWith('/messenger/api') || url.pathname.startsWith('/messenger/ws')) {
    return
  }

  // 네비게이션 요청 (HTML 페이지) - Network first, 오프라인 시 캐시 폴백
  if (request.mode === 'navigate') {
    event.respondWith(
      fetch(request)
        .then((response) => {
          const clone = response.clone()
          caches.open(CACHE_NAME).then((cache) => cache.put(request, clone))
          return response
        })
        .catch(() => {
          return caches.match(request).then((cached) => {
            return cached || caches.match(OFFLINE_URL)
          })
        })
    )
    return
  }

  // 정적 에셋 (JS/CSS/이미지/폰트) - Cache first, network fallback
  if (
    request.destination === 'script' ||
    request.destination === 'style' ||
    request.destination === 'image' ||
    request.destination === 'font'
  ) {
    event.respondWith(
      caches.match(request).then((cached) => {
        if (cached) {
          // 백그라운드에서 업데이트
          fetch(request).then((response) => {
            caches.open(CACHE_NAME).then((cache) => cache.put(request, response))
          }).catch(() => {})
          return cached
        }
        return fetch(request).then((response) => {
          const clone = response.clone()
          caches.open(CACHE_NAME).then((cache) => cache.put(request, clone))
          return response
        })
      })
    )
    return
  }
})

// 푸시 알림 수신 처리
self.addEventListener('push', (event) => {
  if (!event.data) {
    console.log('Push event but no data')
    return
  }

  let data
  try {
    data = event.data.json()
  } catch (e) {
    data = {
      title: 'Mmessenger',
      body: event.data.text(),
      icon: '/messenger/icons/icon-192x192.png'
    }
  }

  const title = data.title || 'Mmessenger'
  const options = {
    body: data.body || '',
    icon: data.icon || '/messenger/icons/icon-192x192.png',
    badge: '/messenger/icons/icon-192x192.png',
    tag: data.tag || 'mmessenger-push',
    renotify: true,
    data: data.data || {}
  }

  event.waitUntil(
    self.registration.showNotification(title, options)
  )
})

// 알림 클릭 처리
self.addEventListener('notificationclick', (event) => {
  event.notification.close()

  const data = event.notification.data || {}
  let url = '/messenger/'

  // 채팅방 ID가 있으면 해당 채팅방으로 이동
  if (data.roomId) {
    url = `/messenger/?room=${data.roomId}`
  }

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true })
      .then((clientList) => {
        // 이미 열려있는 창이 있으면 포커스
        for (const client of clientList) {
          if (client.url.includes(self.location.origin) && 'focus' in client) {
            return client.focus().then((focusedClient) => {
              if ('navigate' in focusedClient) {
                return focusedClient.navigate(url)
              }
            })
          }
        }
        // 열려있는 창이 없으면 새 창 열기
        if (clients.openWindow) {
          return clients.openWindow(url)
        }
      })
  )
})
