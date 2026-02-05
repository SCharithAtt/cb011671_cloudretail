<template>
  <div class="product-details">
    <div v-if="loading" class="loading">Loading product...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    
    <div v-else-if="product" class="product-container">
      <div class="product-main">
        <h1>{{ product.name }}</h1>
        <p class="description">{{ product.description }}</p>
        
        <div class="product-info">
          <span class="price">${{ product.price.toFixed(2) }}</span>
          <span class="stock" :class="{ 'out-of-stock': product.stock <= 0 }">
            {{ product.stock > 0 ? `${product.stock} in stock` : 'Out of stock' }}
          </span>
        </div>

        <div v-if="!authStore.isSeller && product.stock > 0" class="purchase-section">
          <div class="quantity-selector">
            <label>Quantity:</label>
            <input 
              type="number" 
              v-model.number="quantity" 
              min="1" 
              :max="product.stock"
            />
          </div>
          <button @click="addToCart" class="btn btn-primary">
            Add to Cart
          </button>
        </div>
      </div>

      <div class="reviews-section">
        <h2>Reviews</h2>
        
        <div v-if="authStore.isLoggedIn && !authStore.isSeller" class="add-review">
          <h3>Write a Review</h3>
          <textarea 
            v-model="newReview" 
            placeholder="Share your experience..."
            rows="4"
          ></textarea>
          <div class="rating-input">
            <label>Rating:</label>
            <select v-model.number="newRating">
              <option value="5">5 Stars</option>
              <option value="4">4 Stars</option>
              <option value="3">3 Stars</option>
              <option value="2">2 Stars</option>
              <option value="1">1 Star</option>
            </select>
          </div>
          <button @click="submitReview" class="btn btn-secondary">
            Submit Review
          </button>
        </div>

        <div class="reviews-list">
          <div v-if="product.reviews && product.reviews.length > 0">
            <div 
              v-for="(review, index) in product.reviews" 
              :key="index"
              class="review-card"
            >
              <div class="review-header">
                <span class="rating">{{ '★'.repeat(review.rating) }}</span>
                <span class="date">{{ formatDate(review.timestamp) }}</span>
              </div>
              <p>{{ review.comment }}</p>
            </div>
          </div>
          <p v-else class="no-reviews">No reviews yet. Be the first to review!</p>
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

interface Review {
  rating: number
  comment: string
  timestamp: string
}

interface Product {
  product_id: string
  name: string
  description: string
  price: number
  stock: number
  seller_id: string
  reviews?: Review[]
}

const product = ref<Product | null>(null)

const GET_PRODUCT = gql`
  query GetProductById($productId: ID!) {
    getProductById(productId: $productId) {
      product_id
      name
      description
      price
      stock
      seller_id
      reviews {
        rating
        comment
        timestamp
      }
    }
  }
`

const ADD_REVIEW = gql`
  mutation AddReview($productId: ID!, $rating: Int!, $comment: String!) {
    addReview(productId: $productId, rating: $rating, comment: $comment) {
      product_id
      reviews {
        rating
        comment
        timestamp
      }
    }
  }
`

const { result, loading: queryLoading, error: queryError } = useQuery(GET_PRODUCT, {
  productId
})

const { mutate: addReviewMutation } = useMutation(ADD_REVIEW)

onMounted(() => {
  const checkData = setInterval(() => {
    if (!queryLoading.value) {
      if (queryError.value) {
        error.value = 'Failed to load product'
      } else if (result.value) {
        product.value = result.value.getProductById
      }
      loading.value = false
      clearInterval(checkData)
    }
  }, 100)
})

const addToCart = () => {
  if (product.value) {
    cartStore.addItem({
      productId: product.value.product_id,
      quantity: quantity.value,
      name: product.value.name,
      price: product.value.price,
      sellerId: product.value.seller_id
    })
    alert('Added to cart!')
  }
}

const submitReview = async () => {
  if (!newReview.value.trim()) {
    alert('Please write a review')
    return
  }

  try {
    await addReviewMutation({
      productId,
      rating: newRating.value,
      comment: newReview.value
    })

    newReview.value = ''
    newRating.value = 5
    alert('Review submitted!')
    
    // Reload product to get updated reviews
    window.location.reload()
  } catch (err) {
    console.error('Failed to submit review:', err)
    alert('Failed to submit review')
  }
}

const formatDate = (timestamp: string) => {
  return new Date(timestamp).toLocaleDateString()
}
</script>

<style scoped>
.product-details {
  padding: 2rem 0;
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

.product-container {
  display: grid;
  gap: 2rem;
}

.product-main {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

h1 {
  margin-bottom: 1rem;
  color: #2c3e50;
}

.description {
  color: #7f8c8d;
  margin-bottom: 1.5rem;
  line-height: 1.6;
}

.product-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 0;
  border-top: 1px solid #eee;
  border-bottom: 1px solid #eee;
  margin-bottom: 1.5rem;
}

.price {
  font-size: 2rem;
  font-weight: bold;
  color: #2ecc71;
}

.stock {
  color: #27ae60;
  font-weight: 500;
}

.stock.out-of-stock {
  color: #e74c3c;
}

.purchase-section {
  display: flex;
  gap: 1rem;
  align-items: flex-end;
}

.quantity-selector {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.quantity-selector input {
  width: 80px;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.reviews-section {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.add-review {
  margin-bottom: 2rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid #eee;
}

.add-review textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  margin: 1rem 0;
  font-family: inherit;
}

.rating-input {
  margin-bottom: 1rem;
}

.rating-input select {
  margin-left: 0.5rem;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.reviews-list {
  margin-top: 1rem;
}

.review-card {
  padding: 1rem;
  border: 1px solid #eee;
  border-radius: 4px;
  margin-bottom: 1rem;
}

.review-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.rating {
  color: #f39c12;
}

.date {
  color: #7f8c8d;
  font-size: 0.9rem;
}

.no-reviews {
  text-align: center;
  color: #7f8c8d;
  padding: 2rem;
}

.btn {
  padding: 0.75rem 2rem;
  border: none;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background-color: #3498db;
  color: white;
}

.btn-primary:hover {
  background-color: #2980b9;
}

.btn-secondary {
  background-color: #2ecc71;
  color: white;
}

.btn-secondary:hover {
  background-color: #27ae60;
}
</style>
