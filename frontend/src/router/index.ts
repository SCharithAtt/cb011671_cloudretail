import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue')
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { guest: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/RegisterOptions.vue'),
    meta: { guest: true }
  },
  {
    path: '/register/seller',
    name: 'SellerRegister',
    component: () => import('@/views/Register.vue'),
    meta: { guest: true }
  },
  {
    path: '/callback',
    name: 'OAuthCallback',
    component: () => import('@/views/OAuthCallback.vue')
  },
  {
    path: '/products',
    name: 'Products',
    component: () => import('@/views/ProductList.vue')
  },
  {
    path: '/products/:id',
    name: 'ProductDetails',
    component: () => import('@/views/ProductDetails.vue'),
    props: true
  },
  {
    path: '/cart',
    name: 'Cart',
    component: () => import('@/views/Cart.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/checkout',
    name: 'Checkout',
    component: () => import('@/views/Checkout.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/payment/:orderId',
    name: 'Payment',
    component: () => import('@/views/Payment.vue'),
    props: true,
    meta: { requiresAuth: true }
  },
  {
    path: '/order-confirmed/:orderId',
    name: 'OrderConfirmed',
    component: () => import('@/views/OrderConfirmed.vue'),
    props: true,
    meta: { requiresAuth: true }
  },
  {
    path: '/orders',
    name: 'Orders',
    component: () => import('@/views/Orders.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/seller',
    name: 'SellerDashboard',
    component: () => import('@/views/SellerDashboard.vue'),
    meta: { requiresAuth: true, requiresRole: 'seller' }
  },
  {
    path: '/seller/products',
    name: 'SellerProducts',
    component: () => import('@/views/SellerProducts.vue'),
    meta: { requiresAuth: true, requiresRole: 'seller' }
  },
  {
    path: '/seller/products/add',
    name: 'AddProduct',
    component: () => import('@/views/AddProduct.vue'),
    meta: { requiresAuth: true, requiresRole: 'seller' }
  },
  {
    path: '/seller/products/edit/:productId',
    name: 'EditProduct',
    component: () => import('@/views/EditProduct.vue'),
    props: true,
    meta: { requiresAuth: true, requiresRole: 'seller' }
  },
  {
    path: '/seller/orders',
    name: 'SellerOrders',
    component: () => import('@/views/SellerOrders.vue'),
    meta: { requiresAuth: true, requiresRole: 'seller' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guards
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)
  const requiresRole = to.meta.requiresRole as string | undefined
  const isGuest = to.matched.some(record => record.meta.guest)

  // If route is for guests only and user is logged in, redirect to home
  if (isGuest && authStore.isLoggedIn) {
    return next({ name: 'Home' })
  }

  // If route requires authentication and user is not logged in
  if (requiresAuth && !authStore.isLoggedIn) {
    return next({ 
      name: 'Login', 
      query: { redirect: to.fullPath } 
    })
  }

  // If route requires specific role and user doesn't have it
  if (requiresRole && authStore.role !== requiresRole) {
    // Redirect sellers to seller dashboard, buyers to home
    if (authStore.isSeller) {
      return next({ name: 'SellerDashboard' })
    } else {
      return next({ name: 'Home' })
    }
  }

  next()
})

export default router
