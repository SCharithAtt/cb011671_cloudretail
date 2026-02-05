<template>
  <div class="cart">
    <h1>Shopping Cart</h1>
    
    <div v-if="cartStore.isEmpty" class="empty-cart">
      <p>Your cart is empty</p>
      <router-link to="/products" class="btn btn-primary">
        Browse Products
      </router-link>
    </div>

    <div v-else class="cart-container">
      <div class="cart-items">
        <div 
          v-for="item in cartStore.items" 
          :key="item.productId"
          class="cart-item"
        >
          <div class="item-info">
            <h3>{{ item.name }}</h3>
            <p class="price">${{ item.price?.toFixed(2) }}</p>
          </div>
          
          <div class="item-controls">
            <div class="quantity-control">
              <button @click="decreaseQuantity(item.productId)">-</button>
              <span>{{ item.quantity }}</span>
              <button @click="increaseQuantity(item.productId)">+</button>
            </div>
            
            <p class="subtotal">
              ${{ ((item.price || 0) * item.quantity).toFixed(2) }}
            </p>
            
            <button 
              @click="removeItem(item.productId)"
              class="btn-remove"
            >
              Remove
            </button>
          </div>
        </div>
      </div>

      <div class="cart-summary">
        <h2>Order Summary</h2>
        <div class="summary-row">
          <span>Items ({{ cartStore.itemCount }}):</span>
          <span>${{ cartStore.total.toFixed(2) }}</span>
        </div>
        <div class="summary-row total">
          <span>Total:</span>
          <span>${{ cartStore.total.toFixed(2) }}</span>
        </div>
        <router-link to="/checkout" class="btn btn-checkout">
          Proceed to Checkout
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useCartStore } from '@/stores/cart'

const cartStore = useCartStore()

const increaseQuantity = (productId: string) => {
  const item = cartStore.getItem(productId)
  if (item) {
    cartStore.updateQuantity(productId, item.quantity + 1)
  }
}

const decreaseQuantity = (productId: string) => {
  const item = cartStore.getItem(productId)
  if (item) {
    cartStore.updateQuantity(productId, item.quantity - 1)
  }
}

const removeItem = (productId: string) => {
  if (confirm('Remove this item from cart?')) {
    cartStore.removeItem(productId)
  }
}
</script>

<style scoped>
.cart {
  padding: 2rem 0;
}

h1 {
  margin-bottom: 2rem;
  color: #2c3e50;
}

.empty-cart {
  text-align: center;
  padding: 4rem 2rem;
  background: white;
  border-radius: 8px;
}

.empty-cart p {
  font-size: 1.2rem;
  color: #7f8c8d;
  margin-bottom: 1.5rem;
}

.cart-container {
  display: grid;
  grid-template-columns: 1fr 400px;
  gap: 2rem;
}

.cart-items {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.cart-item {
  background: white;
  padding: 1.5rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.item-info h3 {
  margin-bottom: 0.5rem;
  color: #2c3e50;
}

.price {
  color: #2ecc71;
  font-weight: 600;
  font-size: 1.1rem;
}

.item-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.quantity-control {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.quantity-control button {
  width: 30px;
  height: 30px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.quantity-control button:hover {
  background: #3498db;
  color: white;
  border-color: #3498db;
}

.subtotal {
  font-weight: 600;
  color: #2c3e50;
}

.btn-remove {
  padding: 0.5rem 1rem;
  background: #e74c3c;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-remove:hover {
  background: #c0392b;
}

.cart-summary {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  height: fit-content;
}

.cart-summary h2 {
  margin-bottom: 1.5rem;
  color: #2c3e50;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #eee;
}

.summary-row.total {
  font-size: 1.3rem;
  font-weight: bold;
  color: #2c3e50;
  border-bottom: none;
  margin-top: 1rem;
}

.btn-checkout {
  display: block;
  width: 100%;
  padding: 1rem;
  background: #2ecc71;
  color: white;
  text-align: center;
  border-radius: 4px;
  text-decoration: none;
  font-weight: 600;
  margin-top: 1.5rem;
  transition: background 0.2s;
}

.btn-checkout:hover {
  background: #27ae60;
  text-decoration: none;
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

@media (max-width: 768px) {
  .cart-container {
    grid-template-columns: 1fr;
  }
  
  .cart-item {
    grid-template-columns: 1fr;
  }
}
</style>
