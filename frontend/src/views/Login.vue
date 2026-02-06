<template>
  <div class="flex items-center justify-center min-h-[60vh]">
    <div class="card p-8 w-full max-w-md">
      <h2 class="text-2xl font-bold text-center text-gray-900 mb-6">Login</h2>

      <div class="flex gap-2 mb-6">
        <button
          :class="[
            'flex-1 py-3 rounded-lg font-medium transition-all duration-200',
            loginType === 'user' ? 'bg-brand-500 text-white shadow-md' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
          ]"
          @click="loginType = 'user'"
        >Customer</button>
        <button
          :class="[
            'flex-1 py-3 rounded-lg font-medium transition-all duration-200',
            loginType === 'seller' ? 'bg-brand-500 text-white shadow-md' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
          ]"
          @click="loginType = 'seller'"
        >Seller</button>
      </div>

      <form v-if="loginType === 'seller'" @submit.prevent="handleSellerLogin" class="space-y-4">
        <div>
          <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
          <input type="email" id="email" v-model="email" required placeholder="Enter your email" class="input-field" />
        </div>
        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
          <input type="password" id="password" v-model="password" required placeholder="Enter your password" class="input-field" />
        </div>
        <div v-if="error" class="bg-red-50 text-red-600 p-3 rounded-lg text-center text-sm">{{ error }}</div>
        <button type="submit" class="btn-brand w-full" :disabled="loading" :class="{ 'opacity-60 cursor-not-allowed': loading }">
          {{ loading ? 'Logging in...' : 'Login as Seller' }}
        </button>
      </form>

      <div v-else class="text-center py-8">
        <div class="w-16 h-16 bg-brand-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <span class="text-brand-600 text-3xl">&#128274;</span>
        </div>
        <p class="text-gray-500 mb-4">Customer login uses AWS Cognito OAuth2</p>
        <button @click="handleUserLogin" class="btn-brand w-full">Login with Cognito</button>
      </div>

      <p class="text-center mt-6 text-gray-500 text-sm">
        Don't have an account?
        <a @click="handleCognitoSignup" class="text-brand-600 font-medium hover:text-brand-700 cursor-pointer">Register here</a>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const loginType = ref<'user' | 'seller'>('user')
const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

const handleUserLogin = async () => {
  loading.value = true
  error.value = ''
  await authStore.login({ service: 'user', email: email.value, password: password.value })
}

const handleSellerLogin = async () => {
  loading.value = true
  error.value = ''
  const result = await authStore.login({ service: 'seller', email: email.value, password: password.value })
  loading.value = false
  if (result.success) {
    const redirect = route.query.redirect as string
    router.push(redirect || '/seller')
  } else {
    error.value = result.error || 'Login failed'
  }
}

const cognitoDomain = import.meta.env.VITE_COGNITO_DOMAIN || 'https://cloudretail.auth.us-east-1.amazoncognito.com'
const clientId = import.meta.env.VITE_COGNITO_CLIENT_ID || ''
const redirectUri = import.meta.env.VITE_REDIRECT_URI || `${window.location.origin}/callback`

const handleCognitoSignup = () => {
  const params = new URLSearchParams({
    client_id: clientId,
    response_type: 'code',
    scope: 'openid email profile',
    redirect_uri: redirectUri
  })
  window.location.href = `${cognitoDomain}/signup?${params.toString()}`
}
</script>
