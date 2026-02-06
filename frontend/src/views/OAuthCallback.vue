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
  const code = route.query.code as string
  if (code) {
    try {
      await authStore.handleOAuthCallback(code)
      router.push('/')
    } catch (err) {
      error.value = 'Authentication failed. Please try again.'
      setTimeout(() => router.push('/login'), 3000)
    }
  } else {
    error.value = 'No authorization code received.'
    setTimeout(() => router.push('/login'), 3000)
  }
})
</script>
