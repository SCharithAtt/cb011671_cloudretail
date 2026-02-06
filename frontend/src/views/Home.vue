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
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
const authStore = useAuthStore()
</script>
