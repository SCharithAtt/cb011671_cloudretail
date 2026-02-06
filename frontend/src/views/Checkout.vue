<template>
  <div class="py-6 max-w-2xl mx-auto">
    <h1 class="text-3xl font-bold text-gray-900 mb-8">Checkout</h1>

    <div class="card p-6 mb-6">
      <h2 class="text-xl font-semibold text-gray-900 mb-4">Order Items</h2>
      <div class="divide-y divide-gray-100">
        <div v-for="item in cartStore.items" :key="item.productId" class="flex justify-between py-3">
          <div>
            <p class="font-medium text-gray-900">{{ item.name }}</p>
            <p class="text-sm text-gray-500">Qty: {{ item.quantity }}</p>
          </div>
          <p class="font-semibold text-gray-900">${{ ((item.price || 0) * item.quantity).toFixed(2) }}</p>
        </div>
      </div>
      <div class="flex justify-between pt-4 mt-4 border-t border-gray-200 text-lg font-bold">
        <span>Total</span>
        <span class="text-brand-600">${{ cartStore.total.toFixed(2) }}</span>
      </div>
    </div>

    <div v-if="error" class="bg-red-50 text-red-600 p-3 rounded-lg text-center text-sm mb-4">{{ error }}</div>

    <button @click="placeOrder" class="btn-brand w-full text-lg py-4" :disabled="loading" :class="{ 'opacity-60 cursor-not-allowed': loading }">
      {{ loading ? 'Placing Order...' : 'Place Order' }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { orderServiceApi } from '@/services/api'
import { useCartStore } from '@/stores/cart'

const router = useRouter()
const cartStore = useCartStore()
const loading = ref(false)
const error = ref('')

const placeOrder = async () => {
  loading.value = true
  error.value = ''

  try {
    const items = cartStore.items.map(item => ({
      productId: item.productId,
      quantity: item.quantity,
      name: item.name,
      price: item.price,
      sellerId: item.sellerId
    }))

    const response = await orderServiceApi.post('/createOrder', { items })
    const orderId = response.data.orderId

    cartStore.clear()

    if (orderId) {
      router.push(`/payment/${orderId}`)
    } else {
      router.push('/orders')
    }
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to place order'
    loading.value = false
  }
}
</script>
