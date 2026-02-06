<template>
  <div class="py-6 max-w-lg mx-auto">
    <div class="card p-8 text-center">
      <div class="w-20 h-20 bg-brand-100 rounded-full flex items-center justify-center mx-auto mb-6">
        <span class="text-brand-600 text-4xl">&#128179;</span>
      </div>
      <h1 class="text-2xl font-bold text-gray-900 mb-2">Payment</h1>
      <p class="text-gray-500 mb-6">Order #{{ orderId }}</p>

      <div v-if="loading" class="text-gray-500">Processing...</div>
      <div v-else-if="error" class="bg-red-50 text-red-600 p-3 rounded-lg mb-4">{{ error }}</div>

      <div v-if="paymentUrl" class="space-y-4">
        <p class="text-gray-600">Click below to simulate payment:</p>
        <a :href="paymentUrl" class="btn-brand w-full block text-center">Simulate Payment</a>
      </div>
      <div v-else-if="!loading" class="space-y-4">
        <button @click="initiatePayment" class="btn-brand w-full">Initiate Payment</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { orderServiceApi } from '@/services/api'

const route = useRoute()
const orderId = route.params.orderId as string
const loading = ref(true)
const error = ref('')
const paymentUrl = ref('')

const initiatePayment = async () => {
  loading.value = true
  error.value = ''
  try {
    const response = await orderServiceApi.get(`/simulatePayment/${orderId}`)
    paymentUrl.value = response.data.payment_url || ''
    loading.value = false
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to initiate payment'
    loading.value = false
  }
}

onMounted(() => {
  initiatePayment()
})
</script>
