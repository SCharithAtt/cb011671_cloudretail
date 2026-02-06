import axios, { AxiosInstance, InternalAxiosRequestConfig, AxiosError } from 'axios'
import { useAuthStore } from '@/stores/auth'

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_GATEWAY_URL || 'http://localhost:8080',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor to add JWT token
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const authStore = useAuthStore()
    
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for token refresh on 401
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const authStore = useAuthStore()
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    // If 401 and we haven't retried yet, attempt token refresh
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      try {
        const refreshed = await authStore.refreshAccessToken()
        
        if (refreshed && originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${authStore.token}`
          return api(originalRequest)
        }
      } catch (refreshError) {
        // Token refresh failed, logout user
        authStore.logout()
        window.location.href = '/login'
        return Promise.reject(refreshError)
      }
    }

    // If still unauthorized after refresh attempt, logout
    if (error.response?.status === 401) {
      authStore.logout()
      window.location.href = '/login'
    }

    return Promise.reject(error)
  }
)

// Service-specific API instances (all use API Gateway URL in production)
const apiGatewayUrl = import.meta.env.VITE_API_GATEWAY_URL || 'http://localhost:8080'

export const userServiceApi = axios.create({
  baseURL: apiGatewayUrl,
  timeout: 30000
})

export const sellerServiceApi = axios.create({
  baseURL: apiGatewayUrl,
  timeout: 30000
})

export const orderServiceApi = axios.create({
  baseURL: apiGatewayUrl,
  timeout: 30000
})

// Add auth interceptors to service-specific instances
const addAuthInterceptor = (instance: AxiosInstance) => {
  instance.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    const authStore = useAuthStore()
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    return config
  })
}

addAuthInterceptor(userServiceApi)
addAuthInterceptor(sellerServiceApi)
addAuthInterceptor(orderServiceApi)

export default api
