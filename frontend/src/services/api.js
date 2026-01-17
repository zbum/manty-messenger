import axios from 'axios'
import keycloak from './keycloak'

const api = axios.create({
  baseURL: '/messenger/api/v1',
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor
api.interceptors.request.use(
  async (config) => {
    if (keycloak.authenticated) {
      try {
        await keycloak.updateToken(30)
        config.headers.Authorization = `Bearer ${keycloak.token}`
      } catch (error) {
        console.error('Failed to refresh token', error)
      }
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      keycloak.login()
    }
    return Promise.reject(error)
  }
)

// File upload function
export const uploadFile = async (file, onProgress) => {
  const formData = new FormData()
  formData.append('file', file)

  const response = await api.post('/files/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    onUploadProgress: (progressEvent) => {
      if (onProgress && progressEvent.total) {
        const percent = Math.round((progressEvent.loaded * 100) / progressEvent.total)
        onProgress(percent)
      }
    }
  })

  return response.data
}

// Get file type from mime type
export const getFileType = (mimeType) => {
  if (mimeType?.startsWith('image/')) {
    return 'image'
  }
  return 'file'
}

export default api
