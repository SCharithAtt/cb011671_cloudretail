<template>
  <div class="py-6">
    <h1 class="text-3xl font-bold text-gray-900 mb-8">Order Management</h1>

    <div v-if="loading" class="text-center py-12 text-gray-500 text-lg">Loading orders...</div>
    <div v-else-if="error" class="text-center py-12 text-red-500">{{ error }}</div>
    <div v-else-if="orders.length === 0" class="card p-12 text-center">
      <p class="text-gray-500 text-lg">No orders yet.</p>
    </div>

    <div v-else class="space-y-4">
      <div v-for="order in orders" :key="order.orderId" class="card p-6">
        <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3 mb-4">
          <div>
            <p class="font-medium text-gray-900">Order #{{ order.orderId }}</p>
            <p class="text-xs text-gray-400">{{ new Date(order.createdAt).toLocaleString() }}</p>
          </div>
          <div class="flex items-center gap-3">
            <select v-model="order.newStatus" class="input-field w-auto text-sm py-2">
              <option value="pending">Pending</option>
              <option value="confirmed">Confirmed</option>
              <option value="shipped">Shipped</option>
              <option value="delivered">Delivered</option>
              <option value="cancelled">Cancelled</option>
            </select>
            <button @click="updateStatus(order)" class="btn-brand text-sm px-4 py-2">Update</button>
          </div>
        </div>
        <div class="divide-y divide-gray-50">
          <div v-for="item in order.items" :key="item.productId" class="flex justify-between py-2 text-sm">
            <span class="text-gray-700">{{ item.name }} &times; {{ item.quantity }}</span>
            <span class="text-gray-900 font-medium">${{ ((item.price || 0) * item.quantity).toFixed(2) }}</span>
          </div>
        </div>
        <div class="flex justify-between items-center pt-4 mt-4 border-t border-gray-100">
          <span :class="[
            'badge',
            order.status === 'confirmed' ? 'bg-green-100 text-green-700' :
            order.status === 'shipped' ? 'bg-blue-100 text-blue-700' :
            order.status === 'delivered' ? 'bg-purple-100 text-purple-700' :
            order.status === 'cancelled' ? 'bg-red-100 text-red-700' :
            'bg-yellow-100 text-yellow-700'
          ]">{{ order.status }}</span>
          <span class="font-bold text-brand-600 text-lg">${{ order.totalPrice?.toFixed(2) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { sellerServiceApi } from '@/services/api'

const orders = ref<any[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const response = await sellerServiceApi.get('/orders')
    orders.value = (response.data || []).map((o: any) => ({ ...o, newStatus: o.status }))
  } catch (err) {
    error.value = 'Failed to load orders'
  } finally {
    loading.value = false
  }
})

const updateStatus = async (order: any) => {
  try {
    await sellerServiceApi.put(`/updateOrderStatus/${order.orderId}`, { status: order.newStatus })
    order.status = order.newStatus
    alert('Order status updated!')
  } catch (err) {
    alert('Failed to update status')
  }
}
</script>
