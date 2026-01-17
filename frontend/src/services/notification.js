import api from './api'

/**
 * Web Notification & Push Service
 * 브라우저 알림 및 Web Push를 관리하는 서비스
 */

class NotificationService {
  constructor() {
    this.permission = 'default'
    this.isSupported = 'Notification' in window
    this.isPushSupported = 'serviceWorker' in navigator && 'PushManager' in window
    this.swRegistration = null
    this.pushSubscription = null
  }

  /**
   * 서비스 워커 등록
   * @returns {Promise<ServiceWorkerRegistration|null>}
   */
  async registerServiceWorker() {
    if (!this.isPushSupported) {
      console.warn('Push notifications are not supported')
      return null
    }

    try {
      // Vite의 base URL을 사용하여 서비스 워커 경로 설정
      const swPath = `${import.meta.env.BASE_URL}sw.js`
      this.swRegistration = await navigator.serviceWorker.register(swPath)
      console.log('Service Worker registered:', this.swRegistration)
      return this.swRegistration
    } catch (error) {
      console.error('Service Worker registration failed:', error)
      return null
    }
  }

  /**
   * VAPID 공개키 가져오기
   * @returns {Promise<string|null>}
   */
  async getVapidPublicKey() {
    try {
      const response = await api.get('/push/vapid-public-key')
      return response.data.public_key
    } catch (error) {
      console.error('Failed to get VAPID public key:', error)
      return null
    }
  }

  /**
   * URL-safe Base64를 Uint8Array로 변환
   */
  urlBase64ToUint8Array(base64String) {
    const padding = '='.repeat((4 - base64String.length % 4) % 4)
    const base64 = (base64String + padding)
      .replace(/-/g, '+')
      .replace(/_/g, '/')

    const rawData = window.atob(base64)
    const outputArray = new Uint8Array(rawData.length)

    for (let i = 0; i < rawData.length; ++i) {
      outputArray[i] = rawData.charCodeAt(i)
    }
    return outputArray
  }

  /**
   * 푸시 알림 구독
   * @returns {Promise<PushSubscription|null>}
   */
  async subscribePush() {
    if (!this.isPushSupported) {
      console.warn('Push notifications are not supported')
      return null
    }

    // 서비스 워커가 등록되어 있지 않으면 등록
    if (!this.swRegistration) {
      await this.registerServiceWorker()
    }

    if (!this.swRegistration) {
      return null
    }

    try {
      // VAPID 공개키 가져오기
      const vapidPublicKey = await this.getVapidPublicKey()
      if (!vapidPublicKey) {
        console.error('No VAPID public key available')
        return null
      }

      // 기존 구독 확인
      let subscription = await this.swRegistration.pushManager.getSubscription()

      if (!subscription) {
        // 새 구독 생성
        subscription = await this.swRegistration.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: this.urlBase64ToUint8Array(vapidPublicKey)
        })
      }

      this.pushSubscription = subscription

      // 서버에 구독 정보 전송
      await this.sendSubscriptionToServer(subscription)

      console.log('Push subscription successful:', subscription)
      return subscription
    } catch (error) {
      console.error('Failed to subscribe to push notifications:', error)
      return null
    }
  }

  /**
   * 구독 정보를 서버에 전송
   * @param {PushSubscription} subscription
   */
  async sendSubscriptionToServer(subscription) {
    try {
      const subscriptionJSON = subscription.toJSON()
      await api.post('/push/subscribe', {
        endpoint: subscriptionJSON.endpoint,
        keys: {
          p256dh: subscriptionJSON.keys.p256dh,
          auth: subscriptionJSON.keys.auth
        }
      })
      console.log('Subscription sent to server')
    } catch (error) {
      console.error('Failed to send subscription to server:', error)
    }
  }

  /**
   * 푸시 알림 구독 취소
   */
  async unsubscribePush() {
    if (!this.pushSubscription) {
      return
    }

    try {
      // 서버에서 구독 삭제
      await api.delete('/push/unsubscribe')

      // 브라우저에서 구독 취소
      await this.pushSubscription.unsubscribe()
      this.pushSubscription = null

      console.log('Push subscription cancelled')
    } catch (error) {
      console.error('Failed to unsubscribe from push:', error)
    }
  }

  /**
   * 현재 푸시 구독 상태 확인
   * @returns {Promise<boolean>}
   */
  async isPushSubscribed() {
    if (!this.swRegistration) {
      await this.registerServiceWorker()
    }

    if (!this.swRegistration) {
      return false
    }

    const subscription = await this.swRegistration.pushManager.getSubscription()
    return subscription !== null
  }

  /**
   * 알림 권한 상태 확인
   */
  checkPermission() {
    if (!this.isSupported) {
      return 'unsupported'
    }
    this.permission = Notification.permission
    return this.permission
  }

  /**
   * 알림 권한 요청
   * @returns {Promise<string>} 권한 상태 ('granted', 'denied', 'default')
   */
  async requestPermission() {
    if (!this.isSupported) {
      console.warn('This browser does not support notifications')
      return 'unsupported'
    }

    try {
      this.permission = await Notification.requestPermission()
      return this.permission
    } catch (error) {
      console.error('Failed to request notification permission:', error)
      return 'denied'
    }
  }

  /**
   * 알림 표시
   * @param {string} title - 알림 제목
   * @param {Object} options - 알림 옵션
   * @param {Function} onClick - 클릭 콜백
   * @returns {Notification|null}
   */
  show(title, options = {}, onClick = null) {
    if (!this.isSupported || this.permission !== 'granted') {
      return null
    }

    // 페이지가 포커스되어 있으면 알림 표시 안함
    if (document.hasFocus()) {
      return null
    }

    const defaultOptions = {
      icon: '/favicon.ico',
      badge: '/favicon.ico',
      tag: 'mmessenger',
      renotify: true,
      requireInteraction: false,
      silent: false,
      ...options
    }

    try {
      const notification = new Notification(title, defaultOptions)

      if (onClick) {
        notification.onclick = (event) => {
          event.preventDefault()
          window.focus()
          onClick(event)
          notification.close()
        }
      }

      // 5초 후 자동으로 닫기
      setTimeout(() => {
        notification.close()
      }, 5000)

      return notification
    } catch (error) {
      console.error('Failed to show notification:', error)
      return null
    }
  }

  /**
   * 새 메시지 알림 표시
   * @param {Object} message - 메시지 객체
   * @param {Object} room - 채팅방 객체
   * @param {Function} onClickCallback - 클릭 시 호출될 콜백
   */
  showNewMessage(message, room, onClickCallback = null) {
    if (!message || !room) return

    const senderName = message.sender?.display_name || message.sender?.username || 'Unknown'
    const roomName = room.name || 'Chat'

    let body = message.content || ''

    // 메시지 타입에 따른 본문 처리
    if (message.message_type === 'image') {
      body = '📷 이미지를 보냈습니다'
    } else if (message.message_type === 'file') {
      body = '📎 파일을 보냈습니다'
    } else if (body.length > 100) {
      body = body.substring(0, 100) + '...'
    }

    const title = `${senderName} - ${roomName}`

    this.show(title, {
      body,
      tag: `message-${room.id}`,
      data: { roomId: room.id, messageId: message.id }
    }, onClickCallback)
  }

  /**
   * 채팅방 초대 알림 표시
   * @param {Object} room - 초대받은 채팅방
   * @param {Function} onClickCallback - 클릭 시 호출될 콜백
   */
  showRoomInvite(room, onClickCallback = null) {
    if (!room) return

    const title = '채팅방 초대'
    const body = `'${room.name}' 채팅방에 초대되었습니다`

    this.show(title, {
      body,
      tag: `invite-${room.id}`,
      data: { roomId: room.id }
    }, onClickCallback)
  }
}

// 싱글톤 인스턴스
const notificationService = new NotificationService()

export default notificationService
