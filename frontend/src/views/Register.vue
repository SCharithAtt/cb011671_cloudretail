<template>
  <div class="flex items-center justify-center min-h-[60vh]">
    <div class="card p-8 w-full max-w-md">
      <h2 class="text-2xl font-bold text-center text-gray-900 mb-2">Register as Seller</h2>
      <p class="text-center text-gray-500 text-sm mb-6">
        Only seller registration is available here. Customer accounts use AWS Cognito OAuth2.
      </p>

      <form @submit.prevent="handleRegister" class="space-y-4">
        <div>
          <label for="name" class="block text-sm font-medium text-gray-700 mb-1">Business Name</label>
          <input type="text" id="name" v-model="name" required placeholder="Enter your business name" class="input-field" />
        </div>
        <div>
          <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
          <input type="email" id="email" v-model="email" required placeholder="Enter your email" class="input-field" />
        </div>
        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
          <input type="password" id="password" v-model="password" required minlength="8" placeholder="At least 8 characters" class="input-field" />
        </div>
        <div>
          <label for="confirmPassword" class="block text-sm font-medium text-gray-700 mb-1">Confirm Password</label>
          <input type="password" id="confirmPassword" v-model="confirmPassword" required placeholder="Re-enter your password" class="input-field" />
        </div>

        <div v-if="error" class="bg-red-50 text-red-600 p-3 rounded-lg text-center text-sm">{{ error }}</div>
        <div v-if="success" class="bg-green-50 text-green-600 p-3 rounded-lg text-center text-sm">{{ success }}</div>

        <button type="submit" class="btn-brand w-full" :disabled="loading" :class="{ 'opacity-60 cursor-not-allowed': loading }">
          {{ loading ? 'Registering...' : 'Register' }}
        </button>
      </form>

      <p class="text-center mt-6 text-gray-500 text-sm">
        Already have an account?
        <router-link to="/login" class="text-brand-600 font-medium hover:text-brand-700">Login here</router-link>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const name = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')
const success = ref('')

const handleRegister = async () => {
  error.value = ''
  success.value = ''
  if (password.value !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    return
  }
  loading.value = true
  const result = await authStore.registerSeller({ service: 'seller', email: email.value, password: password.value, name: name.value })
  loading.value = false
  if (result.success) {
    success.value = 'Registration successful! Redirecting to login...'
    setTimeout(() => router.push('/login'), 2000)
  } else {
    error.value = result.error || 'Registration failed'
  }
}
</script>
