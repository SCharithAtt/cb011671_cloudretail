import { ApolloClient, InMemoryCache, createHttpLink, ApolloLink } from '@apollo/client/core'
import { setContext } from '@apollo/client/link/context'
import { onError } from '@apollo/client/link/error'
import { useAuthStore } from '@/stores/auth'

// HTTP connection to the GraphQL API
const httpLink = createHttpLink({
  uri: import.meta.env.VITE_GRAPHQL_URL || 'http://localhost:8082/graphql'
})

// Auth link to add JWT token to requests
const authLink = setContext((_, { headers }) => {
  const authStore = useAuthStore()
  const token = authStore.token

  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : ''
    }
  }
})

// Error handling link
const errorLink = onError(({ graphQLErrors, networkError }) => {
  if (graphQLErrors) {
    graphQLErrors.forEach(({ message, locations, path }) => {
      console.error(
        `[GraphQL error]: Message: ${message}, Location: ${locations}, Path: ${path}`
      )
    })
  }

  if (networkError) {
    console.error(`[Network error]: ${networkError}`)
    
    // Handle 401 unauthorized
    if ('statusCode' in networkError && networkError.statusCode === 401) {
      const authStore = useAuthStore()
      authStore.logout()
      window.location.href = '/login'
    }
  }
})

// Create Apollo Client
const apolloClient = new ApolloClient({
  link: ApolloLink.from([errorLink, authLink, httpLink]),
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'network-only',
      errorPolicy: 'all'
    },
    query: {
      fetchPolicy: 'network-only',
      errorPolicy: 'all'
    },
    mutate: {
      errorPolicy: 'all'
    }
  }
})

export default apolloClient
