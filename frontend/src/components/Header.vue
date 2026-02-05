<template>
  <header class="header">
    <div class="container">
      <router-link to="/" class="logo">CloudRetail</router-link>
      
      <nav class="nav">
        <router-link to="/products">Products</router-link>
        
        <template v-if="authStore.isLoggedIn">
          <template v-if="authStore.isSeller">
            <router-link to="/seller">Dashboard</router-link>
            <router-link to="/seller/products">My Products</router-link>
            <router-link to="/seller/orders">Orders</router-link>
          </template>
          <template v-else>
            <router-link to="/cart" class="cart-link">
              Cart
              <span v-if="cartStore.itemCount > 0" class="cart-badge">
                {{ cartStore.itemCount }}
              </span>
            </router-link>
            <router-link to="/orders">My Orders</router-link>
          </template>
          
          <button @click="handleLogout" class="btn-logout">Logout</button>
        </template>
        
        <template v-else>
          <router-link to="/login">Login</router-link>
          <router-link to="/register">Register</router-link>
        </template>
      </nav>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'

const router = useRouter()
const authStore = useAuthStore()
const cartStore = useCartStore()

const handleLogout = () => {
  authStore.logout()
  cartStore.clear()
  router.push('/')
}
</script>

<style scoped>
.header {
  background-color: #2c3e50;
  color: white;
  padding: 1rem 0;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logo {
  font-size: 1.5rem;
  font-weight: bold;
  color: white;
  text-decoration: none;
}

.logo:hover {
  text-decoration: none;
  color: #3498db;
}

.nav {
  display: flex;
  gap: 1.5rem;
  align-items: center;
}

.nav a {
  color: white;
  text-decoration: none;
  transition: color 0.2s;
}

.nav a:hover {
  color: #3498db;
  text-decoration: none;
}

.cart-link {
  position: relative;
}

.cart-badge {
  position: absolute;
  top: -8px;
  right: -12px;
  background-color: #e74c3c;
  color: white;
  border-radius: 10px;
  padding: 2px 6px;
  font-size: 0.75rem;
  font-weight: bold;
}

.btn-logout {
  background-color: transparent;
  color: white;
  border: 1px solid white;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  transition: all 0.2s;
}

.btn-logout:hover {
  background-color: white;
  color: #2c3e50;
}
</style>
