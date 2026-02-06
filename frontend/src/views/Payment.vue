<template>
  <div class="py-6 max-w-lg mx-auto">
    <div class="card p-8 text-center">
      <div class="w-20 h-20 bg-brand-100 rounded-full flex items-center justify-center mx-auto mb-6">
        <span class="text-brand-600 text-4xl">&#128179;</span>
      </div>
      <h1 class="text-2xl font-bold text-gray-900 mb-2">Payment</h1>
      <p class="text-gray-500 mb-2">Order #{{ orderId }}</p>

      <div v-if="loading" class="text-gray-500 py-4">Loading payment info...</div>
      <div v-else-if="error" class="bg-red-50 text-red-600 p-3 rounded-lg mb-4">{{ error }}</div>

      <template v-if="!loading && !error">
        <div class="bg-brand-50 rounded-lg p-4 mb-6">
          <p class="text-sm text-brand-700 font-medium">Total Amount</p>
          <p class="text-3xl font-bold text-brand-800">${{ totalPrice.toFixed(2) }}</p>
          <p class="text-sm text-gray-500 mt-1">Status: {{ orderStatus }}</p>
        </div>

        <div v-if="orderStatus === 'pending'" class="space-y-4">
          <label class="flex items-center gap-3 justify-center cursor-pointer">
            <input type="checkbox" v-model="paid" class="w-5 h-5 accent-brand-500" />
            <span class="text-gray-700 font-medium">I confirm payment</span>
          </label>
          <button @click="markPaymentDone" :disabled="!paid || processing" class="btn-brand w-full" :class="{ 'opacity-60 cursor-not-allowed': !paid || processing }">
            {{ processing ? 'Processing...' : 'Complete Payment' }}
          </button>
        </div>

        <div v-else class="space-y-4">
          <p class="text-green-600 font-semibold">Payment completed!</p>
          <router-link :to="`/order-confirmed/${orderId}`" class="btn-brand w-full block text-center">View Confirmation</router-link>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { orderServiceApi } from '@/services/api'

const route = useRoute()
const router = useRouter()
const orderId = route.params.orderId as string
const loading = ref(true)
const error = ref('')
const processing = ref(false)
const paid = ref(false)
const totalPrice = ref(0)
const orderStatus = ref('')

onMounted(async () => {
  try {
    const response = await orderServiceApi.get(`/simulatePayment/${orderId}`)
    totalPrice.value = response.data.totalPrice || 0
    orderStatus.value = response.data.status || 'pending'
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to load payment info'
  } finally {
    loading.value = false
  }
})

const markPaymentDone = async () => {
  processing.value = true
  try {
    await orderServiceApi.post(`/markPaymentDone/${orderId}`, { paid: true })
    orderStatus.value = 'paid'
    setTimeout(() => router.push(`/order-confirmed/${orderId}`), 1000)
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Payment failed'
  } finally {
    processing.value = false
  }
}
</script>
