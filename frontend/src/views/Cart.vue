<template>
  <div class="py-6">
    <h1 class="text-3xl font-bold text-gray-900 mb-8">Shopping Cart</h1>

    <div v-if="cartStore.isEmpty" class="card p-12 text-center">
      <div class="w-20 h-20 bg-brand-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <span class="text-brand-500 text-4xl">&#128722;</span>
      </div>
      <p class="text-gray-500 text-lg mb-4">Your cart is empty</p>
      <router-link to="/products" class="btn-brand">Browse Products</router-link>
    </div>

    <div v-else class="grid lg:grid-cols-3 gap-8">
      <div class="lg:col-span-2 space-y-4">
        <div v-for="item in cartStore.items" :key="item.productId" class="card p-5 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <h3 class="font-semibold text-gray-900">{{ item.name }}</h3>
            <p class="text-brand-600 font-medium">LKR {{ item.price?.toFixed(2) }}</p>
          </div>
          <div class="flex items-center gap-4">
            <div class="flex items-center gap-2">
              <button @click="decreaseQuantity(item.productId)" class="w-8 h-8 border border-gray-300 rounded-lg flex items-center justify-center hover:bg-brand-500 hover:text-white hover:border-brand-500 transition-all">-</button>
              <span class="w-8 text-center font-medium">{{ item.quantity }}</span>
              <button @click="increaseQuantity(item.productId)" class="w-8 h-8 border border-gray-300 rounded-lg flex items-center justify-center hover:bg-brand-500 hover:text-white hover:border-brand-500 transition-all">+</button>
            </div>
            <p class="font-semibold text-gray-900 w-20 text-right">LKR {{ ((item.price || 0) * item.quantity).toFixed(2) }}</p>
            <button @click="removeItem(item.productId)" class="btn-danger text-sm px-3 py-1.5">Remove</button>
          </div>
        </div>
      </div>

      <div class="card p-6 h-fit sticky top-24">
        <h2 class="text-xl font-bold text-gray-900 mb-4">Order Summary</h2>
        <div class="flex justify-between py-3 border-b border-gray-100 text-gray-600">
          <span>Items ({{ cartStore.itemCount }})</span>
          <span>LKR {{ cartStore.total.toFixed(2) }}</span>
        </div>
        <div class="flex justify-between py-3 text-lg font-bold text-gray-900">
          <span>Total</span>
          <span class="text-brand-600">LKR {{ cartStore.total.toFixed(2) }}</span>
        </div>
        <router-link to="/checkout" class="btn-brand w-full mt-4 text-center block">
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
  if (item) cartStore.updateQuantity(productId, item.quantity + 1)
}

const decreaseQuantity = (productId: string) => {
  const item = cartStore.getItem(productId)
  if (item) cartStore.updateQuantity(productId, item.quantity - 1)
}

const removeItem = (productId: string) => {
  if (confirm('Remove this item from cart?')) cartStore.removeItem(productId)
}
</script>
