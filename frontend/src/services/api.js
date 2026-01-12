import axios from 'axios'

const api = axios.create({
  baseURL: '/messenger/api/v1',
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      const refreshToken = localStorage.getItem('refresh_token')
      if (refreshToken) {
        try {
          const response = await axios.post('/messenger/api/v1/auth/refresh', {
            refresh_token: refreshToken
          })

          const { access_token } = response.data
          localStorage.setItem('access_token', access_token)

          originalRequest.headers.Authorization = `Bearer ${access_token}`
          return api(originalRequest)
        } catch (refreshError) {
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
          window.location.href = '/messenger/login'
        }
      }
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
