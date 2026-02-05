<template>
  <div class="payment">
    <h1>Payment Simulation</h1>
    
    <div v-if="loading" class="loading">Loading payment page...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    
    <div v-else class="payment-container">
      <div class="payment-info">
        <h2>Order ID: {{ orderId }}</h2>
        <p class="info">
          This is a simulated payment page. Check the box below to simulate a successful payment.
        </p>
      </div>

      <div class="payment-form">
        <div class="checkbox-group">
          <input 
            type="checkbox" 
            id="paymentSuccess" 
            v-model="paymentSuccess"
          />
          <label for="paymentSuccess">
            Simulate Successful Payment
          </label>
        </div>

        <button 
          @click="completePayment"
          class="btn btn-primary"
          :disabled="submitting"
        >
          {{ submitting ? 'Processing...' : 'Complete Payment' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { orderServiceApi } from '@/services/api'

const router = useRouter()
const route = useRoute()

const orderId = route.params.orderId as string
const loading = ref(true)
const error = ref('')
const submitting = ref(false)
const paymentSuccess = ref(false)

onMounted(async () => {
  try {
    // Call simulatePayment endpoint to get payment page
    await orderServiceApi.get(`/simulatePayment/${orderId}`)
    loading.value = false
  } catch (err: any) {
    console.error('Failed to load payment page:', err)
    error.value = err.response?.data?.error || 'Failed to load payment page'
    loading.value = false
  }
})

const completePayment = async () => {
  submitting.value = true
  error.value = ''

  try {
    // Mark payment as done
    await orderServiceApi.post(`/markPaymentDone/${orderId}`, {
      paid: paymentSuccess.value
    })

    // Redirect to order confirmed page
    router.push(`/order-confirmed/${orderId}`)
  } catch (err: any) {
    console.error('Payment processing failed:', err)
    error.value = err.response?.data?.error || 'Payment processing failed'
    submitting.value = false
  }
}
</script>

<style scoped>
.payment {
  padding: 2rem 0;
}

h1 {
  margin-bottom: 2rem;
  color: #2c3e50;
  text-align: center;
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
  margin: 0 auto;
}

.payment-container {
  max-width: 600px;
  margin: 0 auto;
}

.payment-info,
.payment-form {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 1.5rem;
}

h2 {
  margin-bottom: 1rem;
  color: #2c3e50;
}

.info {
  color: #7f8c8d;
  line-height: 1.6;
}

.checkbox-group {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1.5rem;
  background: #f8f9fa;
  border-radius: 4px;
  margin-bottom: 1.5rem;
}

.checkbox-group input[type="checkbox"] {
  width: 20px;
  height: 20px;
  cursor: pointer;
}

.checkbox-group label {
  font-size: 1.1rem;
  cursor: pointer;
  color: #2c3e50;
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
  background: #3498db;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2980b9;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
