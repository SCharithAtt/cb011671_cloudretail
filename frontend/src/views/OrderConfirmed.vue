<template>
  <div class="order-confirmed">
    <div v-if="loading" class="loading">Checking order status...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    
    <div v-else-if="orderStatus" class="confirmation-container">
      <div 
        :class="['confirmation-card', orderStatus === 'confirmed' ? 'success' : 'failed']"
      >
        <div class="icon">
          {{ orderStatus === 'confirmed' ? '✓' : '✗' }}
        </div>
        
        <h1>
          {{ orderStatus === 'confirmed' ? 'Order Confirmed!' : 'Payment Failed' }}
        </h1>
        
        <p class="message">
          <template v-if="orderStatus === 'confirmed'">
            Your order has been successfully placed and payment confirmed.
            <br />
            Order ID: <strong>{{ orderId }}</strong>
          </template>
          <template v-else>
            Unfortunately, your payment was not successful. Please try again.
          </template>
        </p>

        <div class="actions">
          <router-link 
            v-if="orderStatus === 'confirmed'" 
            to="/orders" 
            class="btn btn-primary"
          >
            View My Orders
          </router-link>
          <router-link 
            v-else 
            :to="`/payment/${orderId}`"
            class="btn btn-primary"
          >
            Try Again
          </router-link>
          <router-link to="/" class="btn btn-secondary">
            Back to Home
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { orderServiceApi } from '@/services/api'

const route = useRoute()

const orderId = route.params.orderId as string
const loading = ref(true)
const error = ref('')
const orderStatus = ref('')

onMounted(async () => {
  try {
    const response = await orderServiceApi.get(`/orderConfirmed/${orderId}`)
    orderStatus.value = response.data.status
    loading.value = false
  } catch (err: any) {
    console.error('Failed to check order status:', err)
    error.value = err.response?.data?.error || 'Failed to check order status'
    loading.value = false
  }
})
</script>

<style scoped>
.order-confirmed {
  padding: 2rem 0;
  min-height: 60vh;
  display: flex;
  align-items: center;
  justify-content: center;
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

.confirmation-container {
  width: 100%;
  max-width: 600px;
}

.confirmation-card {
  background: white;
  padding: 3rem;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  text-align: center;
}

.confirmation-card.success {
  border-top: 4px solid #2ecc71;
}

.confirmation-card.failed {
  border-top: 4px solid #e74c3c;
}

.icon {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 3rem;
  margin: 0 auto 1.5rem;
  color: white;
}

.success .icon {
  background-color: #2ecc71;
}

.failed .icon {
  background-color: #e74c3c;
}

h1 {
  margin-bottom: 1rem;
  color: #2c3e50;
}

.message {
  color: #7f8c8d;
  line-height: 1.6;
  margin-bottom: 2rem;
}

.message strong {
  color: #2c3e50;
}

.actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
}

.btn {
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  text-decoration: none;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-primary {
  background: #3498db;
  color: white;
}

.btn-primary:hover {
  background: #2980b9;
  text-decoration: none;
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
