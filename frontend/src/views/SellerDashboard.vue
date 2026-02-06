<template>
  <div class="py-6">
    <h1 class="text-3xl font-bold text-gray-900 mb-8">Seller Dashboard</h1>

    <div class="grid md:grid-cols-3 gap-6 mb-10">
      <div class="card p-6 text-center border-t-4 border-brand-400">
        <p class="text-sm font-medium text-gray-500 uppercase tracking-wide mb-2">Products</p>
        <p class="text-4xl font-extrabold text-gray-900 mb-3">{{ stats.totalProducts }}</p>
        <router-link to="/seller/products" class="text-brand-600 font-medium hover:text-brand-700 text-sm">Manage Products &rarr;</router-link>
      </div>
      <div class="card p-6 text-center border-t-4 border-blue-400">
        <p class="text-sm font-medium text-gray-500 uppercase tracking-wide mb-2">Orders</p>
        <p class="text-4xl font-extrabold text-gray-900 mb-3">{{ stats.totalOrders }}</p>
        <router-link to="/seller/orders" class="text-brand-600 font-medium hover:text-brand-700 text-sm">View Orders &rarr;</router-link>
      </div>
      <div class="card p-6 text-center border-t-4 border-green-400">
        <p class="text-sm font-medium text-gray-500 uppercase tracking-wide mb-2">Revenue</p>
        <p class="text-4xl font-extrabold text-gray-900 mb-3">LKR {{ stats.totalRevenue.toFixed(2) }}</p>
        <span class="text-gray-400 text-sm">Total earnings</span>
      </div>
    </div>

    <div class="card p-6">
      <h2 class="text-xl font-bold text-gray-900 mb-4">Quick Actions</h2>
      <router-link to="/seller/products/add" class="btn-brand">
        + Add New Product
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { gql } from '@apollo/client/core'
import { sellerServiceApi } from '@/services/api'

const stats = ref({ totalProducts: 0, totalOrders: 0, totalRevenue: 0 })

const GET_ALL_PRODUCTS = gql`
  query GetAllProducts { getAllProducts { productId sellerId } }
`

const { result } = useQuery(GET_ALL_PRODUCTS)

onMounted(async () => {
  if (result.value) stats.value.totalProducts = result.value.getAllProducts?.length || 0
  try {
    const response = await sellerServiceApi.get('/orders')
    const orders = response.data || []
    stats.value.totalOrders = orders.length
    stats.value.totalRevenue = orders.reduce((sum: number, order: any) => sum + (order.totalPrice || 0), 0)
  } catch (err) {
    console.error('Failed to load stats:', err)
  }
})
</script>
