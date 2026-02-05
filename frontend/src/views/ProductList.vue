<template>
  <div class="product-list">
    <h1>Products</h1>
    
    <div v-if="loading" class="loading">Loading products...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    
    <div v-else class="products-grid">
      <div 
        v-for="product in products" 
        :key="product.product_id"
        class="product-card"
        @click="goToProduct(product.product_id)"
      >
        <div class="product-info">
          <h3>{{ product.name }}</h3>
          <p class="description">{{ product.description }}</p>
          <div class="product-meta">
            <span class="price">${{ product.price.toFixed(2) }}</span>
            <span class="stock" :class="{ 'out-of-stock': product.stock <= 0 }">
              {{ product.stock > 0 ? `${product.stock} in stock` : 'Out of stock' }}
            </span>
          </div>
        </div>
        <button 
          v-if="!authStore.isSeller && product.stock > 0"
          @click.stop="addToCart(product)"
          class="btn-add-cart"
        >
          Add to Cart
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useQuery } from '@vue/apollo-composable'
import { gql } from '@apollo/client/core'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'

const router = useRouter()
const authStore = useAuthStore()
const cartStore = useCartStore()

interface Product {
  product_id: string
  name: string
  description: string
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
  if (result.value) {
    products.value = result.value.getAllProducts || []
    loading.value = false
  }
})

// Watch for query completion
const stopWatch = () => {
  if (!queryLoading.value) {
    if (queryError.value) {
      error.value = 'Failed to load products'
    } else if (result.value) {
      products.value = result.value.getAllProducts || []
    }
    loading.value = false
  }
}

// Simple polling since we can't use watch directly
setInterval(stopWatch, 100)

const goToProduct = (productId: string) => {
  router.push(`/products/${productId}`)
}

const addToCart = (product: Product) => {
  cartStore.addItem({
    productId: product.product_id,
    quantity: 1,
    name: product.name,
    price: product.price,
    sellerId: product.seller_id
  })
  alert('Added to cart!')
}
</script>

<style scoped>
.product-list {
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
}

.products-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1.5rem;
}

.product-card {
  background: white;
  border-radius: 8px;
  padding: 1.5rem;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.product-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.product-info h3 {
  margin-bottom: 0.5rem;
  color: #2c3e50;
}

.description {
  color: #7f8c8d;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.product-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.price {
  font-size: 1.5rem;
  font-weight: bold;
  color: #2ecc71;
}

.stock {
  font-size: 0.9rem;
  color: #27ae60;
}

.stock.out-of-stock {
  color: #e74c3c;
}

.btn-add-cart {
  width: 100%;
  padding: 0.75rem;
  background-color: #3498db;
  color: white;
  border: none;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-add-cart:hover {
  background-color: #2980b9;
}
</style>
