<template>
  <div class="callback">
    <div class="loading">
      <h2>Processing login...</h2>
      <p v-if="error" class="error">{{ error }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const error = ref('')

onMounted(() => {
  // Extract tokens from URL query parameters
  const idToken = route.query.id_token as string
  const accessToken = route.query.access_token as string
  const refreshToken = route.query.refresh_token as string

  if (idToken && accessToken) {
    authStore.handleOAuthCallback(idToken, accessToken, refreshToken)
    router.push('/')
  } else {
    error.value = 'Invalid callback parameters'
    setTimeout(() => {
      router.push('/login')
    }, 3000)
  }
})
</script>

<style scoped>
.callback {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 60vh;
}

.loading {
  text-align: center;
}

.error {
  color: #c33;
  margin-top: 1rem;
}
</style>
