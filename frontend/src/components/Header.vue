<template>
  <header class="bg-gray-900 shadow-lg sticky top-0 z-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex items-center justify-between h-16">
        <router-link to="/" class="flex items-center gap-2 text-xl font-bold text-white hover:text-brand-400 transition-colors">
          <span class="text-brand-400 text-2xl">&#9733;</span>
          CloudRetail
        </router-link>

        <nav class="flex items-center gap-6">
          <router-link to="/products" class="text-gray-300 hover:text-brand-400 transition-colors font-medium">
            Products
          </router-link>

          <template v-if="authStore.isLoggedIn">
            <template v-if="authStore.isSeller">
              <router-link to="/seller" class="text-gray-300 hover:text-brand-400 transition-colors font-medium">Dashboard</router-link>
              <router-link to="/seller/products" class="text-gray-300 hover:text-brand-400 transition-colors font-medium">My Products</router-link>
              <router-link to="/seller/orders" class="text-gray-300 hover:text-brand-400 transition-colors font-medium">Orders</router-link>
            </template>
            <template v-else>
              <router-link to="/cart" class="relative text-gray-300 hover:text-brand-400 transition-colors font-medium">
                Cart
                <span v-if="cartStore.itemCount > 0" class="absolute -top-2 -right-4 bg-brand-500 text-white text-xs font-bold rounded-full h-5 w-5 flex items-center justify-center">
                  {{ cartStore.itemCount }}
                </span>
              </router-link>
              <router-link to="/orders" class="text-gray-300 hover:text-brand-400 transition-colors font-medium">My Orders</router-link>
            </template>

            <button @click="handleLogout" class="ml-2 px-4 py-2 border border-brand-500 text-brand-400 rounded-lg hover:bg-brand-500 hover:text-white transition-all duration-200 font-medium text-sm">
              Logout
            </button>
          </template>

          <template v-else>
            <router-link to="/login" class="text-gray-300 hover:text-brand-400 transition-colors font-medium">Login</router-link>
            <router-link to="/register" class="px-4 py-2 bg-brand-500 text-white rounded-lg hover:bg-brand-600 transition-colors font-medium text-sm">Register</router-link>
          </template>
        </nav>
      </div>
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
