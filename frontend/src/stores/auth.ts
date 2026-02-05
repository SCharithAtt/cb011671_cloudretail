import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { jwtDecode } from 'jwt-decode'
import axios from 'axios'

interface JWTPayload {
  sub: string
  email: string
  'custom:role'?: string
  exp: number
}

interface LoginCredentials {
  service: 'user' | 'seller'
  email: string
  password: string
  name?: string // For seller registration
}

export const useAuthStore = defineStore('auth', () => {
  // State
  const token = ref<string>(localStorage.getItem('token') || '')
  const refreshToken = ref<string>(localStorage.getItem('refreshToken') || '')
  const role = ref<string>(localStorage.getItem('role') || '')
  const userId = ref<string>(localStorage.getItem('userId') || '')
  const userEmail = ref<string>(localStorage.getItem('userEmail') || '')

  // Computed
  const isLoggedIn = computed(() => !!token.value && !isTokenExpired())
  const isSeller = computed(() => role.value === 'seller')
  const isBuyer = computed(() => role.value !== 'seller')

  // Actions
  function parseJWT(jwtToken: string) {
    try {
      const decoded = jwtDecode<JWTPayload>(jwtToken)
      userId.value = decoded.sub
      userEmail.value = decoded.email
      role.value = decoded['custom:role'] || ''
      
      // Persist to localStorage
      localStorage.setItem('userId', userId.value)
      localStorage.setItem('userEmail', userEmail.value)
      localStorage.setItem('role', role.value)
      
      return decoded
    } catch (error) {
      console.error('Failed to parse JWT:', error)
      return null
    }
  }

  function isTokenExpired(): boolean {
    if (!token.value) return true
    
    try {
      const decoded = jwtDecode<JWTPayload>(token.value)
      const currentTime = Date.now() / 1000
      return decoded.exp < currentTime
    } catch {
      return true
    }
  }

  async function login(credentials: LoginCredentials) {
    const baseUrl = credentials.service === 'seller' 
      ? import.meta.env.VITE_SELLER_SERVICE_URL || 'http://localhost:8081'
      : import.meta.env.VITE_USER_SERVICE_URL || 'http://localhost:8080'

    try {
      if (credentials.service === 'seller') {
        // Seller login via SellerService
        const response = await axios.post(`${baseUrl}/sellerLogin`, {
          email: credentials.email,
          password: credentials.password
        })

        const { id_token, access_token, refresh_token } = response.data

        token.value = id_token
        refreshToken.value = refresh_token || ''
        
        // Persist to localStorage
        localStorage.setItem('token', token.value)
        localStorage.setItem('refreshToken', refreshToken.value)
        
        // Parse JWT to extract claims
        parseJWT(id_token)
        
        return { success: true, data: response.data }
      } else {
        // User login via UserService (Cognito OAuth2)
        // Redirect to /login endpoint which will redirect to Cognito
        window.location.href = `${baseUrl}/login`
        return { success: true, data: { redirect: true } }
      }
    } catch (error: any) {
      console.error('Login failed:', error)
      return { 
        success: false, 
        error: error.response?.data?.error || 'Login failed' 
      }
    }
  }

  async function registerSeller(credentials: Required<LoginCredentials>) {
    const baseUrl = import.meta.env.VITE_SELLER_SERVICE_URL || 'http://localhost:8081'

    try {
      const response = await axios.post(`${baseUrl}/sellerRegister`, {
        email: credentials.email,
        password: credentials.password,
        name: credentials.name
      })

      return { success: true, data: response.data }
    } catch (error: any) {
      console.error('Registration failed:', error)
      return { 
        success: false, 
        error: error.response?.data?.error || 'Registration failed' 
      }
    }
  }

  function handleOAuthCallback(idToken: string, accessToken: string, refreshTokenParam?: string) {
    // Called after OAuth2 callback from UserService
    token.value = idToken
    refreshToken.value = refreshTokenParam || ''
    
    localStorage.setItem('token', token.value)
    if (refreshTokenParam) {
      localStorage.setItem('refreshToken', refreshTokenParam)
    }
    
    parseJWT(idToken)
  }

  function logout() {
    token.value = ''
    refreshToken.value = ''
    role.value = ''
    userId.value = ''
    userEmail.value = ''
    
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('role')
    localStorage.removeItem('userId')
    localStorage.removeItem('userEmail')
  }

  async function refreshAccessToken() {
    if (!refreshToken.value) {
      logout()
      return false
    }

    try {
      // Attempt to refresh token (implementation depends on Cognito setup)
      // For now, we'll just logout if token is expired
      if (isTokenExpired()) {
        logout()
        return false
      }
      return true
    } catch (error) {
      console.error('Token refresh failed:', error)
      logout()
      return false
    }
  }

  return {
    // State
    token,
    refreshToken,
    role,
    userId,
    userEmail,
    
    // Computed
    isLoggedIn,
    isSeller,
    isBuyer,
    
    // Actions
    login,
    registerSeller,
    logout,
    parseJWT,
    isTokenExpired,
    refreshAccessToken,
    handleOAuthCallback
  }
})
