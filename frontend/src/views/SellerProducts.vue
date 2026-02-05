<template>
  <div class="seller-products">
    <div class="header">
      <h1>My Products</h1>
      <router-link to="/seller/products/add" class="btn btn-primary">
        Add Product
      </router-link>
    </div>
    
    <div v-if="loading" class="loading">Loading products...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    
    <div v-else-if="products.length === 0" class="no-products">
      <p>You haven't added any products yet</p>
      <router-link to="/seller/products/add" class="btn btn-primary">
        Add Your First Product
      </router-link>
    </div>

    <div v-else class="products-table">
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Price</th>
            <th>Stock</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="product in products" :key="product.product_id">
            <td>{{ product.name }}</td>
            <td>${{ product.price.toFixed(2) }}</td>
            <td :class="{ 'low-stock': product.stock < 10 }">
              {{ product.stock }}
            </td>
            <td class="actions">
              <router-link 
                :to="`/seller/products/edit/${product.product_id}`"
                class="btn-edit"
              >
                Edit
              </router-link>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { gql } from '@apollo/client/core'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

interface Product {
  product_id: string
  name: string
  price: number
  stock: number
  seller_id: string
}

const products = ref<Product[]>([])
const loading = ref(true)
const error = ref('')

const GET_ALL_PRODUCTS = gql`
  query GetAllProducts {
    getAllProducts {
      product_id
      name
      description
      price
      stock
      seller_id
    }
  }
`

const { result, loading: queryLoading, error: queryError } = useQuery(GET_ALL_PRODUCTS)

onMounted(() => {
  const checkData = setInterval(() => {
    if (!queryLoading.value) {
      if (queryError.value) {
        error.value = 'Failed to load products'
      } else if (result.value) {
        // Filter products by current seller
        const allProducts = result.value.getAllProducts || []
        products.value = allProducts.filter((p: Product) => p.seller_id === authStore.userId)
      }
      loading.value = false
      clearInterval(checkData)
    }
  }, 100)
})
</script>

<style scoped>
.seller-products {
  padding: 2rem 0;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

h1 {
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
}

.no-products {
  text-align: center;
  padding: 4rem 2rem;
  background: white;
  border-radius: 8px;
}

.no-products p {
  font-size: 1.2rem;
  color: #7f8c8d;
  margin-bottom: 1.5rem;
}

.products-table {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

table {
  width: 100%;
  border-collapse: collapse;
}

thead {
  background: #f8f9fa;
}

th {
  padding: 1rem;
  text-align: left;
  font-weight: 600;
  color: #2c3e50;
  border-bottom: 2px solid #eee;
}

td {
  padding: 1rem;
  border-bottom: 1px solid #eee;
}

.low-stock {
  color: #e74c3c;
  font-weight: 600;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

.btn-edit {
  padding: 0.5rem 1rem;
  background: #3498db;
  color: white;
  text-decoration: none;
  border-radius: 4px;
  font-size: 0.9rem;
  transition: background 0.2s;
}

.btn-edit:hover {
  background: #2980b9;
  text-decoration: none;
}

.btn-primary {
  padding: 0.75rem 1.5rem;
  background: #2ecc71;
  color: white;
  text-decoration: none;
  border-radius: 4px;
  font-weight: 600;
  transition: background 0.2s;
}

.btn-primary:hover {
  background: #27ae60;
  text-decoration: none;
}
</style>
