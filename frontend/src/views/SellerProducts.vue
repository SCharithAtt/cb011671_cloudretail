<template>
  <div class="py-6">
    <div class="flex items-center justify-between mb-8">
      <h1 class="text-3xl font-bold text-gray-900">My Products</h1>
      <router-link to="/seller/products/add" class="btn-brand">+ Add Product</router-link>
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-500 text-lg">Loading products...</div>
    <div v-else-if="error" class="text-center py-12 text-red-500">{{ error }}</div>
    <div v-else-if="products.length === 0" class="card p-12 text-center">
      <p class="text-gray-500 text-lg mb-4">You haven't added any products yet.</p>
      <router-link to="/seller/products/add" class="btn-brand">Add Your First Product</router-link>
    </div>

    <div v-else class="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
      <div v-for="product in products" :key="product.product_id" class="card overflow-hidden">
        <div class="h-2 bg-gradient-to-r from-brand-400 to-brand-600"></div>
        <div class="p-5">
          <h3 class="text-lg font-semibold text-gray-900 mb-1">{{ product.name }}</h3>
          <p class="text-gray-500 text-sm mb-4 line-clamp-2">{{ product.description }}</p>
          <div class="flex justify-between items-center mb-4">
            <span class="text-xl font-bold text-brand-600">${{ product.price.toFixed(2) }}</span>
            <span :class="['badge', product.stock > 0 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700']">
              Stock: {{ product.stock }}
            </span>
          </div>
          <router-link :to="`/seller/products/edit/${product.product_id}`" class="btn-brand-outline w-full text-center block text-sm py-2">
            Edit Product
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { gql } from '@apollo/client/core'

interface Product { product_id: string; name: string; description: string; price: number; stock: number; seller_id: string }

const products = ref<Product[]>([])
const loading = ref(true)
const error = ref('')

const GET_ALL_PRODUCTS = gql`
  query GetAllProducts { getAllProducts { product_id name description price stock seller_id } }
`

const { result, loading: queryLoading, error: queryError } = useQuery(GET_ALL_PRODUCTS)

onMounted(() => {
  const checkData = setInterval(() => {
    if (!queryLoading.value) {
      if (queryError.value) error.value = 'Failed to load products'
      else if (result.value) products.value = result.value.getAllProducts || []
      loading.value = false
      clearInterval(checkData)
    }
  }, 100)
})
</script>
