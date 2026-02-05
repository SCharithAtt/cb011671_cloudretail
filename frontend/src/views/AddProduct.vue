<template>
  <div class="add-product">
    <h1>Add New Product</h1>
    
    <div class="form-container">
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
            {{ submitting ? 'Adding Product...' : 'Add Product' }}
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
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { sellerServiceApi } from '@/services/api'

const router = useRouter()

const form = ref({
  name: '',
  description: '',
  price: 0,
  stock: 0
})

const submitting = ref(false)
const error = ref('')
const success = ref('')

const handleSubmit = async () => {
  submitting.value = true
  error.value = ''
  success.value = ''

  try {
    await sellerServiceApi.post('/addProduct', form.value)
    
    success.value = 'Product added successfully!'
    
    setTimeout(() => {
      router.push('/seller/products')
    }, 1500)
  } catch (err: any) {
    console.error('Failed to add product:', err)
    error.value = err.response?.data?.error || 'Failed to add product'
    submitting.value = false
  }
}
</script>

<style scoped>
.add-product {
  padding: 2rem 0;
}

h1 {
  margin-bottom: 2rem;
  color: #2c3e50;
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

.error {
  background-color: #fee;
  color: #c33;
  padding: 0.75rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  text-align: center;
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
  background: #2ecc71;
  color: white;
  flex: 1;
}

.btn-primary:hover:not(:disabled) {
  background: #27ae60;
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
