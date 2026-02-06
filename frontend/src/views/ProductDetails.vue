<template>
  <div class="py-6">
    <div v-if="loading" class="text-center py-12 text-gray-500 text-lg">Loading product...</div>
    <div v-else-if="error" class="text-center py-12 text-red-500 text-lg">{{ error }}</div>

    <div v-else-if="product" class="space-y-8">
      <div class="card p-8">
        <h1 class="text-3xl font-bold text-gray-900 mb-3">{{ product.name }}</h1>
        <p class="text-gray-500 leading-relaxed mb-6">{{ product.description }}</p>
        <div class="flex items-center justify-between py-4 border-y border-gray-100 mb-6">
          <span class="text-3xl font-bold text-brand-600">${{ product.price.toFixed(2) }}</span>
          <span :class="['badge text-sm px-3 py-1', product.stock > 0 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700']">
            {{ product.stock > 0 ? `${product.stock} in stock` : 'Out of stock' }}
          </span>
        </div>
        <div v-if="!authStore.isSeller && product.stock > 0" class="flex items-end gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Quantity</label>
            <input type="number" v-model.number="quantity" min="1" :max="product.stock" class="input-field w-24" />
          </div>
          <button @click="addToCart" class="btn-brand">Add to Cart</button>
        </div>
      </div>

      <div class="card p-8">
        <h2 class="text-2xl font-bold text-gray-900 mb-6">Reviews</h2>
        <div v-if="authStore.isLoggedIn && !authStore.isSeller" class="pb-6 mb-6 border-b border-gray-100">
          <h3 class="text-lg font-semibold text-gray-800 mb-3">Write a Review</h3>
          <textarea v-model="newReview" placeholder="Share your experience..." rows="4" class="input-field mb-3"></textarea>
          <div class="flex items-center gap-4 mb-4">
            <label class="text-sm font-medium text-gray-700">Rating:</label>
            <select v-model.number="newRating" class="input-field w-auto">
              <option value="5">5 Stars</option>
              <option value="4">4 Stars</option>
              <option value="3">3 Stars</option>
              <option value="2">2 Stars</option>
              <option value="1">1 Star</option>
            </select>
          </div>
          <button @click="submitReview" class="btn-brand">Submit Review</button>
        </div>
        <div class="space-y-4">
          <template v-if="product.reviews && product.reviews.length > 0">
            <div v-for="(review, index) in product.reviews" :key="index" class="p-4 bg-gray-50 rounded-lg">
              <div class="flex justify-between items-center mb-2">
                <span class="text-brand-500 text-lg">{{ '\u2605'.repeat(review.rating) }}<span class="text-gray-300">{{ '\u2605'.repeat(5 - review.rating) }}</span></span>
                <span class="text-gray-400 text-sm">{{ formatDate(review.timestamp) }}</span>
              </div>
              <p class="text-gray-700">{{ review.comment }}</p>
            </div>
          </template>
          <p v-else class="text-center text-gray-400 py-8">No reviews yet. Be the first to review!</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { gql } from '@apollo/client/core'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'

const route = useRoute()
const authStore = useAuthStore()
const cartStore = useCartStore()

const productId = route.params.id as string
const quantity = ref(1)
const newReview = ref('')
const newRating = ref(5)
const loading = ref(true)
const error = ref('')

interface Review { rating: number; comment: string; timestamp: string }
interface Product { product_id: string; name: string; description: string; price: number; stock: number; seller_id: string; reviews?: Review[] }

const product = ref<Product | null>(null)

const GET_PRODUCT = gql`
  query GetProductById($productId: ID!) {
    getProductById(productId: $productId) { product_id name description price stock seller_id reviews { rating comment timestamp } }
  }
`

const ADD_REVIEW = gql`
  mutation AddReview($productId: ID!, $rating: Int!, $comment: String!) {
    addReview(productId: $productId, rating: $rating, comment: $comment) { product_id reviews { rating comment timestamp } }
  }
`

const { result, loading: queryLoading, error: queryError } = useQuery(GET_PRODUCT, { productId })
const { mutate: addReviewMutation } = useMutation(ADD_REVIEW)

onMounted(() => {
  const checkData = setInterval(() => {
    if (!queryLoading.value) {
      if (queryError.value) error.value = 'Failed to load product'
      else if (result.value) product.value = result.value.getProductById
      loading.value = false
      clearInterval(checkData)
    }
  }, 100)
})

const addToCart = () => {
  if (product.value) {
    cartStore.addItem({ productId: product.value.product_id, quantity: quantity.value, name: product.value.name, price: product.value.price, sellerId: product.value.seller_id })
    alert('Added to cart!')
  }
}

const submitReview = async () => {
  if (!newReview.value.trim()) { alert('Please write a review'); return }
  try {
    await addReviewMutation({ productId, rating: newRating.value, comment: newReview.value })
    newReview.value = ''; newRating.value = 5
    alert('Review submitted!'); window.location.reload()
  } catch { alert('Failed to submit review') }
}

const formatDate = (timestamp: string) => new Date(timestamp).toLocaleDateString()
</script>
