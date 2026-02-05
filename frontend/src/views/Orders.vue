<template>
  <div class="orders">
    <h1>My Orders</h1>
    
    <div v-if="loading" class="loading">Loading orders...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    
    <div v-else-if="orders.length === 0" class="no-orders">
      <p>You haven't placed any orders yet</p>
      <router-link to="/products" class="btn btn-primary">
        Start Shopping
      </router-link>
    </div>

    <div v-else class="orders-list">
      <div 
        v-for="order in orders" 
        :key="order.order_id"
        class="order-card"
      >
        <div class="order-header">
          <div>
            <h3>Order #{{ order.order_id.substring(0, 8) }}</h3>
            <p class="date">{{ formatDate(order.created_at) }}</p>
          </div>
          <span :class="['status', order.status]">
            {{ formatStatus(order.status) }}
          </span>
        </div>

        <div class="order-items">
          <div 
            v-for="(item, index) in order.items" 
            :key="index"
            class="order-item"
          >
            <span>{{ item.product_id.substring(0, 8) }}...</span>
            <span>Qty: {{ item.quantity }}</span>
            <span>${{ item.price.toFixed(2) }}</span>
          </div>
        </div>

        <div class="order-total">
          <span>Total:</span>
          <span>${{ order.total_price.toFixed(2) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { orderServiceApi } from '@/services/api'

interface OrderItem {
  product_id: string
  quantity: number
  price: number
  seller_id: string
}

interface Order {
  order_id: string
  buyer_id: string
  seller_id: string
  status: string
  items: OrderItem[]
  total_price: number
  created_at: string
}

const orders = ref<Order[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const response = await orderServiceApi.get('/getOrders')
    orders.value = response.data.orders || []
    loading.value = false
  } catch (err: any) {
    console.error('Failed to load orders:', err)
    error.value = err.response?.data?.error || 'Failed to load orders'
    loading.value = false
  }
})

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

const formatStatus = (status: string) => {
  return status.replace('_', ' ').toUpperCase()
}
</script>

<style scoped>
.orders {
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

.no-orders {
  text-align: center;
  padding: 4rem 2rem;
  background: white;
  border-radius: 8px;
}

.no-orders p {
  font-size: 1.2rem;
  color: #7f8c8d;
  margin-bottom: 1.5rem;
}

.orders-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.order-card {
  background: white;
  padding: 1.5rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #eee;
}

.order-header h3 {
  margin-bottom: 0.25rem;
  color: #2c3e50;
}

.date {
  color: #7f8c8d;
  font-size: 0.9rem;
}

.status {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  font-weight: 600;
  font-size: 0.9rem;
}

.status.pending {
  background: #fff3cd;
  color: #856404;
}

.status.confirmed {
  background: #d4edda;
  color: #155724;
}

.status.shipped {
  background: #d1ecf1;
  color: #0c5460;
}

.status.delivered {
  background: #d4edda;
  color: #155724;
}

.status.cancelled,
.status.not_confirmed {
  background: #f8d7da;
  color: #721c24;
}

.order-items {
  margin: 1rem 0;
}

.order-item {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid #f5f5f5;
}

.order-total {
  display: flex;
  justify-content: space-between;
  padding-top: 1rem;
  margin-top: 1rem;
  border-top: 2px solid #eee;
  font-size: 1.2rem;
  font-weight: bold;
  color: #2c3e50;
}

.btn-primary {
  padding: 1rem 2rem;
  background: #3498db;
  color: white;
  text-decoration: none;
  border-radius: 4px;
  display: inline-block;
  transition: background 0.2s;
}

.btn-primary:hover {
  background: #2980b9;
  text-decoration: none;
}
</style>
