/**
 * Service Worker for Web Push Notifications
 */

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
      icon: '/favicon.ico'
    }
  }

  const title = data.title || 'Mmessenger'
  const options = {
    body: data.body || '',
    icon: data.icon || '/favicon.ico',
    badge: '/favicon.ico',
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
  let url = '/'

  // 채팅방 ID가 있으면 해당 채팅방으로 이동
  if (data.roomId) {
    url = `/?room=${data.roomId}`
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

// 서비스 워커 설치
self.addEventListener('install', (event) => {
  console.log('Service Worker installed')
  self.skipWaiting()
})

// 서비스 워커 활성화
self.addEventListener('activate', (event) => {
  console.log('Service Worker activated')
  event.waitUntil(clients.claim())
})
