<template>
  <div class="py-6 max-w-2xl mx-auto">
    <h1 class="text-3xl font-bold text-gray-900 mb-8">Edit Product</h1>

    <div v-if="loading" class="text-center py-12 text-gray-500 text-lg">Loading product...</div>

    <div v-else class="card p-8">
      <form @submit.prevent="handleSubmit" class="space-y-5">
        <div>
          <label for="name" class="block text-sm font-medium text-gray-700 mb-1">Product Name</label>
          <input type="text" id="name" v-model="form.name" required class="input-field" />
        </div>
        <div>
          <label for="description" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
          <textarea id="description" v-model="form.description" required rows="4" class="input-field"></textarea>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label for="price" class="block text-sm font-medium text-gray-700 mb-1">Price ($)</label>
            <input type="number" id="price" v-model.number="form.price" required min="0" step="0.01" class="input-field" />
          </div>
          <div>
            <label for="stock" class="block text-sm font-medium text-gray-700 mb-1">Stock</label>
            <input type="number" id="stock" v-model.number="form.stock" required min="0" class="input-field" />
          </div>
        </div>

        <div v-if="error" class="bg-red-50 text-red-600 p-3 rounded-lg text-center text-sm">{{ error }}</div>
        <div v-if="success" class="bg-green-50 text-green-600 p-3 rounded-lg text-center text-sm">{{ success }}</div>

        <button type="submit" class="btn-brand w-full" :disabled="saving" :class="{ 'opacity-60 cursor-not-allowed': saving }">
          {{ saving ? 'Saving...' : 'Save Changes' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuery } from '@vue/apollo-composable'
import { gql } from '@apollo/client/core'
import { sellerServiceApi } from '@/services/api'

const route = useRoute()
const router = useRouter()
const productId = route.params.productId as string

const loading = ref(true)
const saving = ref(false)
const error = ref('')
const success = ref('')

const form = ref({ name: '', description: '', price: 0, stock: 0 })

const GET_PRODUCT = gql`
  query GetProductById($id: ID!) {
    getProductById(id: $id) { productId name description price stock imageUrl }
  }
`

const { result, loading: queryLoading } = useQuery(GET_PRODUCT, { id: productId })

onMounted(() => {
  const checkData = setInterval(() => {
    if (!queryLoading.value) {
      if (result.value?.getProductById) {
        const p = result.value.getProductById
        form.value = { name: p.name, description: p.description, price: p.price, stock: p.stock }
      }
      loading.value = false
      clearInterval(checkData)
    }
  }, 100)
})

const handleSubmit = async () => {
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    await sellerServiceApi.put(`/editProduct/${productId}`, form.value)
    success.value = 'Product updated successfully!'
    setTimeout(() => router.push('/seller/products'), 1500)
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to update product'
  } finally {
    saving.value = false
  }
}
</script>
