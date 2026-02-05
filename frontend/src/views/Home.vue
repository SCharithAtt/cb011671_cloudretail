<template>
  <div class="home">
    <h1>Welcome to CloudRetail</h1>
    <p class="subtitle">Your one-stop shop for all your needs</p>
    
    <div class="actions">
      <router-link to="/products" class="btn btn-primary">
        Browse Products
      </router-link>
      <router-link v-if="!authStore.isLoggedIn" to="/login" class="btn btn-secondary">
        Get Started
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
</script>

<style scoped>
.home {
  text-align: center;
  padding: 4rem 2rem;
}

h1 {
  font-size: 3rem;
  margin-bottom: 1rem;
  color: #2c3e50;
}

.subtitle {
  font-size: 1.5rem;
  color: #7f8c8d;
  margin-bottom: 3rem;
}

.actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
}

.btn {
  padding: 1rem 2rem;
  border-radius: 4px;
  text-decoration: none;
  font-size: 1.1rem;
  transition: all 0.2s;
}

.btn-primary {
  background-color: #3498db;
  color: white;
}

.btn-primary:hover {
  background-color: #2980b9;
  text-decoration: none;
}

.btn-secondary {
  background-color: #2ecc71;
  color: white;
}

.btn-secondary:hover {
  background-color: #27ae60;
  text-decoration: none;
}
</style>
