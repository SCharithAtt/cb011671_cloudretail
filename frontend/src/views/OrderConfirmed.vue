<template>
  <div class="py-12 max-w-lg mx-auto text-center">
    <div class="card p-10">
      <div class="w-24 h-24 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-6">
        <span class="text-green-500 text-5xl">&#10003;</span>
      </div>
      <h1 class="text-3xl font-bold text-gray-900 mb-2">Order Confirmed!</h1>
      <p class="text-gray-500 mb-8">Your order has been placed successfully.</p>

      <div v-if="orderId" class="bg-brand-50 rounded-lg p-4 mb-8">
        <p class="text-sm text-brand-700 font-medium">Order ID</p>
        <p class="text-lg font-mono font-bold text-brand-800">{{ orderId }}</p>
      </div>

      <div class="flex gap-4 justify-center">
        <router-link to="/orders" class="btn-brand">View My Orders</router-link>
        <router-link to="/products" class="btn-brand-outline">Continue Shopping</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
const orderId = route.params.orderId as string
</script>
