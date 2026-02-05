import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface CartItem {
  productId: string
  quantity: number
  name?: string
  price?: number
  sellerId?: string
}

export const useCartStore = defineStore('cart', () => {
  // State
  const items = ref<CartItem[]>(loadCartFromStorage())

  // Computed
  const itemCount = computed(() => {
    return items.value.reduce((total, item) => total + item.quantity, 0)
  })

  const total = computed(() => {
    return items.value.reduce((total, item) => {
      return total + (item.price || 0) * item.quantity
    }, 0)
  })

  const isEmpty = computed(() => items.value.length === 0)

  // Actions
  function loadCartFromStorage(): CartItem[] {
    try {
      const stored = localStorage.getItem('cart')
      return stored ? JSON.parse(stored) : []
    } catch {
      return []
    }
  }

  function saveCartToStorage() {
    localStorage.setItem('cart', JSON.stringify(items.value))
  }

  function addItem(product: CartItem) {
    const existingItem = items.value.find(item => item.productId === product.productId)
    
    if (existingItem) {
      existingItem.quantity += product.quantity
    } else {
      items.value.push({ ...product })
    }
    
    saveCartToStorage()
  }

  function updateQuantity(productId: string, quantity: number) {
    const item = items.value.find(item => item.productId === productId)
    
    if (item) {
      if (quantity <= 0) {
        removeItem(productId)
      } else {
        item.quantity = quantity
        saveCartToStorage()
      }
    }
  }

  function removeItem(productId: string) {
    const index = items.value.findIndex(item => item.productId === productId)
    
    if (index !== -1) {
      items.value.splice(index, 1)
      saveCartToStorage()
    }
  }

  function clear() {
    items.value = []
    localStorage.removeItem('cart')
  }

  function getTotal(): number {
    return total.value
  }

  function getItem(productId: string): CartItem | undefined {
    return items.value.find(item => item.productId === productId)
  }

  return {
    // State
    items,
    
    // Computed
    itemCount,
    total,
    isEmpty,
    
    // Actions
    addItem,
    updateQuantity,
    removeItem,
    clear,
    getTotal,
    getItem
  }
})
