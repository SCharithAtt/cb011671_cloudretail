<template>
  <div class="edit-product">
    <h1>Edit Product</h1>
    
    <div v-if="loading" class="loading">Loading product...</div>
    <div v-else-if="loadError" class="error">{{ loadError }}</div>
    
    <div v-else class="form-container">
      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label for="name">Product Name *</label>
          <input 
            type="text" 
            id="name" 
            v-model="form.name" 
            required 
            placeholder="Enter product name"
          />
        </div>

        <div class="form-group">
          <label for="description">Description *</label>
          <textarea 
            id="description" 
            v-model="form.description" 
            required 
            rows="4"
            placeholder="Describe your product"
          ></textarea>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label for="price">Price ($) *</label>
            <input 
              type="number" 
              id="price" 
              v-model.number="form.price" 
              required 
              min="0.01"
              step="0.01"
              placeholder="0.00"
            />
          </div>

          <div class="form-group">
            <label for="stock">Stock *</label>
            <input 
              type="number" 
              id="stock" 
              v-model.number="form.stock" 
              required 
              min="0"
              placeholder="0"
            />
          </div>
        </div>

        <div v-if="error" class="error">{{ error }}</div>
        <div v-if="success" class="success">{{ success }}</div>

        <div class="form-actions">
          <button type="submit" class="btn btn-primary" :disabled="submitting">
            {{ submitting ? 'Updating Product...' : 'Update Product' }}
          </button>
          <router-link to="/seller/products" class="btn btn-secondary">
            Cancel
          </router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useQuery } from '@vue/apollo-composable'
import { gql } from '@apollo/client/core'
import { sellerServiceApi } from '@/services/api'

const router = useRouter()
const route = useRoute()
const productId = route.params.id as string

const form = ref({
  name: '',
  description: '',
  price: 0,
  stock: 0
})

const loading = ref(true)
const loadError = ref('')
const submitting = ref(false)
const error = ref('')
const success = ref('')

const GET_PRODUCT = gql`
  query GetProductById($productId: ID!) {
    getProductById(productId: $productId) {
      product_id
      name
      description
      price
      stock
    }
  }
`

const { result, loading: queryLoading, error: queryError } = useQuery(GET_PRODUCT, {
  productId
})

onMounted(() => {
  const checkData = setInterval(() => {
    if (!queryLoading.value) {
      if (queryError.value) {
        loadError.value = 'Failed to load product'
      } else if (result.value) {
        const product = result.value.getProductById
        form.value = {
          name: product.name,
          description: product.description,
          price: product.price,
          stock: product.stock
        }
      }
      loading.value = false
      clearInterval(checkData)
    }
  }, 100)
})

const handleSubmit = async () => {
  submitting.value = true
  error.value = ''
  success.value = ''

  try {
    await sellerServiceApi.put(`/editProduct/${productId}`, form.value)
    
    success.value = 'Product updated successfully!'
    
    setTimeout(() => {
      router.push('/seller/products')
    }, 1500)
  } catch (err: any) {
    console.error('Failed to update product:', err)
    error.value = err.response?.data?.error || 'Failed to update product'
    submitting.value = false
  }
}
</script>

<style scoped>
.edit-product {
  padding: 2rem 0;
}

h1 {
  margin-bottom: 2rem;
  color: #2c3e50;
}

.loading,
.error {
  text-align: center;
  padding: 2rem;
  font-size: 1.2rem;
}

.error {
  color: #c33;
  background: #fee;
  border-radius: 8px;
  max-width: 600px;
}

.form-container {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  max-width: 600px;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: #2c3e50;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
  font-family: inherit;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #3498db;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.success {
  background-color: #efe;
  color: #2a0;
  padding: 0.75rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  text-align: center;
}

.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 4px;
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  text-align: center;
  transition: all 0.2s;
}

.btn-primary {
  background: #3498db;
  color: white;
  flex: 1;
}

.btn-primary:hover:not(:disabled) {
  background: #2980b9;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: #95a5a6;
  color: white;
}

.btn-secondary:hover {
  background: #7f8c8d;
  text-decoration: none;
}
</style>
