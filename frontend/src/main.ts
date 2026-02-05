import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { DefaultApolloClient } from '@vue/apollo-composable'
import App from './App.vue'
import router from './router'
import apolloClient from './services/graphql'
import './style.css'

const app = createApp(App)
const pinia = createPinia()

// Provide Apollo Client to the app
app.provide(DefaultApolloClient, apolloClient)

// Use plugins
app.use(pinia)
app.use(router)

// Global error handler
app.config.errorHandler = (err, instance, info) => {
  console.error('Global error:', err)
  console.error('Error info:', info)
  
  // You can add error tracking service here (e.g., Sentry)
}

// Mount app
app.mount('#app')
