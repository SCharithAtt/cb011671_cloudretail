<template>
  <div class="seller-dashboard">
    <h1>Seller Dashboard</h1>
    
    <div class="dashboard-cards">
      <div class="card">
        <h3>Products</h3>
        <p class="number">{{ stats.totalProducts }}</p>
        <router-link to="/seller/products">Manage Products</router-link>
      </div>

      <div class="card">
        <h3>Orders</h3>
        <p class="number">{{ stats.totalOrders }}</p>
        <router-link to="/seller/orders">View Orders</router-link>
      </div>

      <div class="card">
        <h3>Revenue</h3>
        <p class="number">${{ stats.totalRevenue.toFixed(2) }}</p>
        <span class="subtitle">Total earnings</span>
      </div>
    </div>

    <div class="quick-actions">
      <h2>Quick Actions</h2>
      <router-link to="/seller/products/add" class="btn btn-primary">
        Add New Product
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { gql } from '@apollo/client/core'
import { sellerServiceApi } from '@/services/api'

const stats = ref({
  totalProducts: 0,
  totalOrders: 0,
  totalRevenue: 0
})

const GET_ALL_PRODUCTS = gql`
  query GetAllProducts {
    getAllProducts {
      product_id
      seller_id
    }
  }
`

const { result } = useQuery(GET_ALL_PRODUCTS)

onMounted(async () => {
  // Get products count (filter by seller_id would happen on server in real app)
  if (result.value) {
    stats.value.totalProducts = result.value.getAllProducts?.length || 0
  }

  // Get orders stats
  try {
    const response = await sellerServiceApi.get('/orders')
    const orders = response.data.orders || []
    stats.value.totalOrders = orders.length
    stats.value.totalRevenue = orders.reduce((sum: number, order: any) => {
      return sum + (order.total_price || 0)
    }, 0)
  } catch (err) {
    console.error('Failed to load stats:', err)
  }
})
</script>

<style scoped>
.seller-dashboard {
  padding: 2rem 0;
}

h1 {
  margin-bottom: 2rem;
  color: #2c3e50;
}

.dashboard-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-bottom: 3rem;
}

.card {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  text-align: center;
}

.card h3 {
  margin-bottom: 1rem;
  color: #7f8c8d;
  font-size: 1rem;
  font-weight: 500;
  text-transform: uppercase;
}

.card .number {
  font-size: 2.5rem;
  font-weight: bold;
  color: #2c3e50;
  margin-bottom: 0.5rem;
}

.card a {
  color: #3498db;
  font-weight: 500;
  transition: color 0.2s;
}

.card a:hover {
  color: #2980b9;
}

.subtitle {
  color: #7f8c8d;
  font-size: 0.9rem;
}

.quick-actions {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.quick-actions h2 {
  margin-bottom: 1rem;
  color: #2c3e50;
}

.btn-primary {
  padding: 0.75rem 1.5rem;
  background: #2ecc71;
  color: white;
  text-decoration: none;
  border-radius: 4px;
  display: inline-block;
  font-weight: 600;
  transition: background 0.2s;
}

.btn-primary:hover {
  background: #27ae60;
  text-decoration: none;
}
</style>
