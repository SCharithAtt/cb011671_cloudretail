<template>
  <div class="py-12">
    <!-- Hero Section -->
    <div class="text-center mb-16">
      <div class="inline-flex items-center gap-2 bg-brand-100 text-brand-700 px-4 py-2 rounded-full text-sm font-medium mb-6">
        <span>&#9733;</span> Welcome to CloudRetail
      </div>
      <h1 class="text-5xl md:text-6xl font-extrabold text-gray-900 mb-6 leading-tight">
        Shop Smarter,<br />
        <span class="text-brand-500">Live Better</span>
      </h1>
      <p class="text-xl text-gray-500 max-w-2xl mx-auto mb-10">
        Discover amazing products from sellers around the world. Quality goods, great prices, and a seamless shopping experience.
      </p>
      <div class="flex gap-4 justify-center">
        <router-link to="/products" class="btn-brand text-lg px-8 py-4">
          Browse Products
        </router-link>
        <router-link v-if="!authStore.isLoggedIn" to="/register" class="btn-brand-outline text-lg px-8 py-4">
          Get Started
        </router-link>
      </div>
    </div>

    <!-- Feature Cards -->
    <div class="grid md:grid-cols-3 gap-8 max-w-5xl mx-auto">
      <div class="card p-8 text-center">
        <div class="w-14 h-14 bg-brand-100 rounded-xl flex items-center justify-center mx-auto mb-4">
          <span class="text-brand-600 text-2xl">&#128722;</span>
        </div>
        <h3 class="text-lg font-semibold text-gray-900 mb-2">Easy Shopping</h3>
        <p class="text-gray-500 text-sm">Browse, add to cart, and checkout in just a few clicks.</p>
      </div>
      <div class="card p-8 text-center">
        <div class="w-14 h-14 bg-brand-100 rounded-xl flex items-center justify-center mx-auto mb-4">
          <span class="text-brand-600 text-2xl">&#128176;</span>
        </div>
        <h3 class="text-lg font-semibold text-gray-900 mb-2">Best Prices</h3>
        <p class="text-gray-500 text-sm">Competitive prices from verified sellers worldwide.</p>
      </div>
      <div class="card p-8 text-center">
        <div class="w-14 h-14 bg-brand-100 rounded-xl flex items-center justify-center mx-auto mb-4">
          <span class="text-brand-600 text-2xl">&#128274;</span>
        </div>
        <h3 class="text-lg font-semibold text-gray-900 mb-2">Secure Payments</h3>
        <p class="text-gray-500 text-sm">Your transactions are protected with bank-level security.</p>
      </div>
    </div>

    <!-- Categories Section -->
    <div class="mt-20 mb-16">
      <h2 class="text-3xl font-bold text-gray-900 text-center mb-10">Shop by Category</h2>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 max-w-5xl mx-auto">
        <router-link to="/products" class="card p-6 text-center hover:border-brand-300 transition-all group">
          <div class="text-4xl mb-3">💻</div>
          <h3 class="font-semibold text-gray-900 group-hover:text-brand-600">Electronics</h3>
        </router-link>
        <router-link to="/products" class="card p-6 text-center hover:border-brand-300 transition-all group">
          <div class="text-4xl mb-3">🏠</div>
          <h3 class="font-semibold text-gray-900 group-hover:text-brand-600">Home & Office</h3>
        </router-link>
        <router-link to="/products" class="card p-6 text-center hover:border-brand-300 transition-all group">
          <div class="text-4xl mb-3">🎮</div>
          <h3 class="font-semibold text-gray-900 group-hover:text-brand-600">Gaming</h3>
        </router-link>
        <router-link to="/products" class="card p-6 text-center hover:border-brand-300 transition-all group">
          <div class="text-4xl mb-3">📱</div>
          <h3 class="font-semibold text-gray-900 group-hover:text-brand-600">Accessories</h3>
        </router-link>
      </div>
    </div>

    <!-- Featured Products Section -->
    <div class="mt-20">
      <h2 class="text-3xl font-bold text-gray-900 text-center mb-10">Featured Products</h2>
      <div v-if="loadingProducts" class="text-center py-8 text-gray-500">Loading products...</div>
      <div v-else-if="featuredProducts.length > 0" class="grid sm:grid-cols-2 lg:grid-cols-4 gap-6 max-w-7xl mx-auto">
        <router-link 
          v-for="product in featuredProducts" 
          :key="product.productId" 
          :to="`/products/${product.productId}`"
          class="card overflow-hidden hover:border-brand-300 transition-all group"
        >
          <img 
            :src="product.imageUrl || 'https://via.placeholder.com/400x300/6366f1/ffffff?text=No+Image'" 
            :alt="product.name"
            class="w-full h-40 object-cover"
          />
          <div class="p-4">
            <h3 class="font-semibold text-gray-900 mb-2 line-clamp-1 group-hover:text-brand-600">{{ product.name }}</h3>
            <p class="text-brand-600 font-bold text-lg">LKR {{ product.price.toFixed(2) }}</p>
          </div>
        </router-link>
      </div>
      <div class="text-center mt-8">
        <router-link to="/products" class="btn-brand-outline">View All Products</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { gql } from '@apollo/client/core'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const loadingProducts = ref(true)
const featuredProducts = ref<any[]>([])

const GET_ALL_PRODUCTS = gql`
  query GetAllProducts {
    getAllProducts {
      productId
      name
      price
      imageUrl
    }
  }
`

const { result } = useQuery(GET_ALL_PRODUCTS)

onMounted(() => {
  // Wait for products to load and pick 4 random ones
  const checkProducts = setInterval(() => {
    if (result.value?.getAllProducts) {
      const allProducts = result.value.getAllProducts
      // Shuffle and take first 4
      const shuffled = [...allProducts].sort(() => Math.random() - 0.5)
      featuredProducts.value = shuffled.slice(0, 4)
      loadingProducts.value = false
      clearInterval(checkProducts)
    }
  }, 100)
})
</script>
