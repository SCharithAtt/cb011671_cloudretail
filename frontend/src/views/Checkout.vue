<template>
  <div class="checkout">
    <h1>Checkout</h1>
    
    <div v-if="loading" class="loading">Processing order...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    
    <div v-else class="checkout-container">
      <div class="order-items">
        <h2>Order Items</h2>
        <div 
          v-for="item in cartStore.items" 
          :key="item.productId"
          class="order-item"
        >
          <span>{{ item.name }} x {{ item.quantity }}</span>
          <span>${{ ((item.price || 0) * item.quantity).toFixed(2) }}</span>
        </div>
        
        <div class="order-total">
          <span>Total:</span>
          <span>${{ cartStore.total.toFixed(2) }}</span>
        </div>
      </div>

      <div class="checkout-form">
        <h2>Confirm Order</h2>
        <p class="info">
          Click the button below to create your order. You will be redirected to the payment page.
        </p>
        
        <button 
          @click="createOrder" 
          class="btn btn-primary"
          :disabled="submitting"
        >
          {{ submitting ? 'Creating Order...' : 'Place Order' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '@/stores/cart'
import { orderServiceApi } from '@/services/api'

const router = useRouter()
const cartStore = useCartStore()

const loading = ref(false)
const error = ref('')
const submitting = ref(false)

const createOrder = async () => {
  if (cartStore.isEmpty) {
    error.value = 'Your cart is empty'
    return
  }

  submitting.value = true
  error.value = ''

  try {
    // Create order via OrderService
    const response = await orderServiceApi.post('/createOrder', {
      items: cartStore.items.map(item => ({
        product_id: item.productId,
        quantity: item.quantity,
        price: item.price,
        seller_id: item.sellerId
      }))
    })

    const { order_id } = response.data

    // Clear cart
    cartStore.clear()

    // Redirect to payment page
    router.push(`/payment/${order_id}`)
  } catch (err: any) {
    console.error('Failed to create order:', err)
    error.value = err.response?.data?.error || 'Failed to create order'
    submitting.value = false
  }
}
</script>

<style scoped>
.checkout {
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
}

.checkout-container {
  display: grid;
  grid-template-columns: 1fr 400px;
  gap: 2rem;
}

.order-items,
.checkout-form {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

h2 {
  margin-bottom: 1.5rem;
  color: #2c3e50;
}

.order-item {
  display: flex;
  justify-content: space-between;
  padding: 1rem 0;
  border-bottom: 1px solid #eee;
}

.order-total {
  display: flex;
  justify-content: space-between;
  padding-top: 1rem;
  margin-top: 1rem;
  border-top: 2px solid #ddd;
  font-size: 1.3rem;
  font-weight: bold;
  color: #2c3e50;
}

.info {
  color: #7f8c8d;
  margin-bottom: 1.5rem;
  line-height: 1.6;
}

.btn {
  width: 100%;
  padding: 1rem;
  border: none;
  border-radius: 4px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #2ecc71;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #27ae60;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .checkout-container {
    grid-template-columns: 1fr;
  }
}
</style>
