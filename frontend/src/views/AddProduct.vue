<template>
  <div class="py-6 max-w-2xl mx-auto">
    <h1 class="text-3xl font-bold text-gray-900 mb-8">Add New Product</h1>

    <div class="card p-8">
      <form @submit.prevent="handleSubmit" class="space-y-5">
        <div>
          <label for="name" class="block text-sm font-medium text-gray-700 mb-1">Product Name</label>
          <input type="text" id="name" v-model="form.name" required class="input-field" placeholder="Enter product name" />
        </div>
        <div>
          <label for="description" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
          <textarea id="description" v-model="form.description" required rows="4" class="input-field" placeholder="Describe your product"></textarea>
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

        <button type="submit" class="btn-brand w-full" :disabled="loading" :class="{ 'opacity-60 cursor-not-allowed': loading }">
          {{ loading ? 'Adding...' : 'Add Product' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { sellerServiceApi } from '@/services/api'

const router = useRouter()
const loading = ref(false)
const error = ref('')
const success = ref('')

const form = ref({ name: '', description: '', price: 0, stock: 0 })

const handleSubmit = async () => {
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    await sellerServiceApi.post('/addProduct', form.value)
    success.value = 'Product added successfully!'
    setTimeout(() => router.push('/seller/products'), 1500)
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to add product'
  } finally {
    loading.value = false
  }
}
</script>
