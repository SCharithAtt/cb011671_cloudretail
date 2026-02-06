<template>
  <div class="flex items-center justify-center min-h-[60vh]">
    <div class="card p-8 w-full max-w-md text-center">
      <h2 class="text-2xl font-bold text-gray-900 mb-2">Create Your Account</h2>
      <p class="text-gray-500 text-sm mb-8">Join CloudRetail today</p>

      <!-- Primary: Customer Registration with Cognito -->
      <button 
        @click="handleCognitoSignup" 
        class="btn-brand w-full py-4 text-lg font-semibold mb-6"
      >
        Register as Customer
      </button>

      <div class="relative mb-6">
        <div class="absolute inset-0 flex items-center">
          <div class="w-full border-t border-gray-200"></div>
        </div>
        <div class="relative flex justify-center">
          <span class="bg-white px-4 text-sm text-gray-400">or</span>
        </div>
      </div>

      <!-- Secondary: Seller Registration Link -->
      <p class="text-gray-500 text-sm">
        Want to sell on CloudRetail?
        <router-link 
          to="/register/seller" 
          class="text-brand-600 font-medium hover:text-brand-700 hover:underline"
        >
          Register as a Seller
        </router-link>
      </p>

      <p class="text-center mt-8 text-gray-500 text-sm">
        Already have an account?
        <router-link to="/login" class="text-brand-600 font-medium hover:text-brand-700">Login here</router-link>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
const cognitoDomain = import.meta.env.VITE_COGNITO_DOMAIN || 'https://cloudretail.auth.us-east-1.amazoncognito.com'
const clientId = import.meta.env.VITE_COGNITO_CLIENT_ID || ''
const redirectUri = import.meta.env.VITE_REDIRECT_URI || `${window.location.origin}/callback`

const handleCognitoSignup = () => {
  const params = new URLSearchParams({
    client_id: clientId,
    response_type: 'code',
    scope: 'openid email profile',
    redirect_uri: redirectUri
  })
  window.location.href = `${cognitoDomain}/signup?${params.toString()}`
}
</script>
