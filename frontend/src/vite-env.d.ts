/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_GATEWAY_URL: string
  readonly VITE_USER_SERVICE_URL: string
  readonly VITE_SELLER_SERVICE_URL: string
  readonly VITE_PRODUCT_SERVICE_URL: string
  readonly VITE_ORDER_SERVICE_URL: string
  readonly VITE_GRAPHQL_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}
