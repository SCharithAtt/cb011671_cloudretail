<template>
  <div class="login">
    <div class="login-card">
      <h2>Login</h2>
      
      <div class="tab-buttons">
        <button 
          :class="{ active: loginType === 'user' }"
          @click="loginType = 'user'"
        >
          Customer
        </button>
        <button 
          :class="{ active: loginType === 'seller' }"
          @click="loginType = 'seller'"
        >
          Seller
        </button>
      </div>

      <form v-if="loginType === 'seller'" @submit.prevent="handleSellerLogin">
        <div class="form-group">
          <label for="email">Email</label>
          <input 
            type="email" 
            id="email" 
            v-model="email" 
            required 
            placeholder="Enter your email"
          />
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <input 
            type="password" 
            id="password" 
            v-model="password" 
            required 
            placeholder="Enter your password"
          />
        </div>

        <div v-if="error" class="error">{{ error }}</div>

        <button type="submit" class="btn btn-primary" :disabled="loading">
          {{ loading ? 'Logging in...' : 'Login as Seller' }}
        </button>
      </form>

      <div v-else class="oauth-login">
        <p>Customer login uses AWS Cognito OAuth2</p>
        <button @click="handleUserLogin" class="btn btn-primary">
          Login with Cognito
        </button>
      </div>

      <p class="switch-text">
        Don't have an account? 
        <router-link to="/register">Register here</router-link>
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
  
  await authStore.login({
    service: 'user',
    email: email.value,
    password: password.value
  })
  
  // UserService will redirect to Cognito
}

const handleSellerLogin = async () => {
  loading.value = true
  error.value = ''

  const result = await authStore.login({
    service: 'seller',
    email: email.value,
    password: password.value
  })

  loading.value = false

  if (result.success) {
    const redirect = route.query.redirect as string
    router.push(redirect || '/seller')
  } else {
    error.value = result.error || 'Login failed'
  }
}
</script>

<style scoped>
.login {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 60vh;
}

.login-card {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 400px;
}

h2 {
  text-align: center;
  margin-bottom: 1.5rem;
  color: #2c3e50;
}

.tab-buttons {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
}

.tab-buttons button {
  flex: 1;
  padding: 0.75rem;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  transition: all 0.2s;
}

.tab-buttons button.active {
  background-color: #3498db;
  color: white;
  border-color: #3498db;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #2c3e50;
}

.form-group input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
}

.form-group input:focus {
  outline: none;
  border-color: #3498db;
}

.oauth-login {
  text-align: center;
  padding: 2rem 0;
}

.oauth-login p {
  margin-bottom: 1rem;
  color: #7f8c8d;
}

.error {
  background-color: #fee;
  color: #c33;
  padding: 0.75rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  text-align: center;
}

.btn {
  width: 100%;
  padding: 0.75rem;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background-color: #3498db;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #2980b9;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.switch-text {
  text-align: center;
  margin-top: 1.5rem;
  color: #7f8c8d;
}

.switch-text a {
  color: #3498db;
  font-weight: 500;
}
</style>
