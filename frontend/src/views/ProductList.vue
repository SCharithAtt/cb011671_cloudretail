<template>
  <div class="py-6">
    <h1 class="text-3xl font-bold text-gray-900 mb-8">Products</h1>

    <div v-if="loading" class="text-center py-12 text-gray-500 text-lg">Loading products...</div>
    <div v-else-if="error" class="text-center py-12 text-red-500 text-lg">{{ error }}</div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
      <div
        v-for="product in products"
        :key="product.productId"
        class="card overflow-hidden cursor-pointer group hover:border-brand-200"
        @click="goToProduct(product.productId)"
      >
        <div class="h-2 bg-gradient-to-r from-brand-400 to-brand-500"></div>
        <div class="p-5">
          <h3 class="text-lg font-semibold text-gray-900 mb-2 group-hover:text-brand-600 transition-colors">{{ product.name }}</h3>
          <p class="text-gray-500 text-sm mb-4 line-clamp-2">{{ product.description }}</p>
          <div class="flex justify-between items-center mb-4">
            <span class="text-2xl font-bold text-brand-600">${{ product.price.toFixed(2) }}</span>
            <span
              :class="[
                'badge',
                product.stock > 0 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
              ]"
            >
              {{ product.stock > 0 ? `${product.stock} in stock` : 'Out of stock' }}
            </span>
          </div>
          <button
            v-if="!authStore.isSeller && product.stock > 0"
            @click.stop="addToCart(product)"
            class="w-full py-2.5 bg-brand-500 text-white rounded-lg font-medium hover:bg-brand-600 transition-colors"
          >
            Add to Cart
          </button>
        </div>
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
  productId: string
  name: string
  description: string
  price: number
  stock: number
  sellerId: string
}

const products = ref<Product[]>([])
const loading = ref(true)
const error = ref('')

const GET_ALL_PRODUCTS = gql`
  query GetAllProducts {
    getAllProducts {
      productId
      name
      description
      price
      stock
      sellerId
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

const stopWatch = () => {
  if (!queryLoading.value) {
    if (queryError.value) error.value = 'Failed to load products'
    else if (result.value) products.value = result.value.getAllProducts || []
    loading.value = false
  }
}

setInterval(stopWatch, 100)

const goToProduct = (productId: string) => router.push(`/products/${productId}`)

const addToCart = (product: Product) => {
  cartStore.addItem({ productId: product.productId, quantity: 1, name: product.name, price: product.price, sellerId: product.sellerId })
  alert('Added to cart!')
}
</script>
