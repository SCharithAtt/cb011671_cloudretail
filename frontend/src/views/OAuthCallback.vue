<template>
  <div class="flex items-center justify-center min-h-[60vh]">
    <div class="card p-10 text-center max-w-sm">
      <div class="w-16 h-16 bg-brand-100 rounded-full flex items-center justify-center mx-auto mb-4 animate-pulse">
        <span class="text-brand-600 text-2xl">&#8987;</span>
      </div>
      <h2 class="text-xl font-semibold text-gray-900 mb-2">Authenticating...</h2>
      <p class="text-gray-500">Please wait while we complete your login.</p>
      <div v-if="error" class="mt-4 bg-red-50 text-red-600 p-3 rounded-lg text-sm">{{ error }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const error = ref('')

onMounted(async () => {
  // Check for error from backend
  const errorParam = route.query.error as string
  if (errorParam) {
    const errorDesc = route.query.error_description as string
    error.value = errorDesc || errorParam || 'Authentication failed.'
    setTimeout(() => router.push('/login'), 3000)
    return
  }

  // Read tokens passed by user_service callback redirect
  const idToken = route.query.id_token as string
  const accessToken = route.query.access_token as string
  const refreshTokenParam = route.query.refresh_token as string

  if (idToken) {
    try {
      authStore.handleOAuthCallback(idToken, accessToken, refreshTokenParam)
      // Clean URL by replacing with home
      router.replace('/')
    } catch (err) {
      error.value = 'Authentication failed. Please try again.'
      setTimeout(() => router.push('/login'), 3000)
    }
  } else {
    // If no tokens, might be a direct Cognito redirect with code
    // (for local dev where user_service isn't in the path)
    const code = route.query.code as string
    if (code) {
      error.value = 'Unexpected authorization code. Login flow may be misconfigured.'
    } else {
      error.value = 'No authentication tokens received.'
    }
    setTimeout(() => router.push('/login'), 3000)
  }
})
</script>
