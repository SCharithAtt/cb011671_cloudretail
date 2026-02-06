<template>
  <div class="py-6">
    <h1 class="text-3xl font-bold text-gray-900 mb-8">My Orders</h1>

    <div v-if="loading" class="text-center py-12 text-gray-500 text-lg">Loading orders...</div>
    <div v-else-if="error" class="text-center py-12 text-red-500">{{ error }}</div>
    <div v-else-if="orders.length === 0" class="card p-12 text-center">
      <p class="text-gray-500 text-lg mb-4">You haven't placed any orders yet.</p>
      <router-link to="/products" class="btn-brand">Start Shopping</router-link>
    </div>

    <div v-else class="space-y-4">
      <div v-for="order in orders" :key="order.orderId" class="card p-6">
        <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3 mb-4">
          <div>
            <p class="text-sm text-gray-500">Order #{{ order.orderId }}</p>
            <p class="text-xs text-gray-400">{{ new Date(order.createdAt).toLocaleString() }}</p>
          </div>
          <span :class="[
            'badge',
            order.status === 'confirmed' ? 'bg-green-100 text-green-700' :
            order.status === 'pending' ? 'bg-yellow-100 text-yellow-700' :
            order.status === 'shipped' ? 'bg-blue-100 text-blue-700' :
            'bg-gray-100 text-gray-700'
          ]">
            {{ order.status }}
          </span>
        </div>
        <div class="divide-y divide-gray-50">
          <div v-for="item in order.items" :key="item.productId" class="flex justify-between py-2 text-sm">
            <span class="text-gray-700">{{ item.name }} &times; {{ item.quantity }}</span>
            <span class="text-gray-900 font-medium">${{ ((item.price || 0) * item.quantity).toFixed(2) }}</span>
          </div>
        </div>
        <div class="flex justify-between items-center pt-4 mt-4 border-t border-gray-100">
          <span class="font-bold text-gray-900">Total</span>
          <span class="font-bold text-brand-600 text-lg">LKR {{ order.totalPrice?.toFixed(2) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { orderServiceApi } from '@/services/api'

const orders = ref<any[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const response = await orderServiceApi.get('/getOrders')
    orders.value = response.data || []
  } catch (err) {
    error.value = 'Failed to load orders'
  } finally {
    loading.value = false
  }
})
</script>
