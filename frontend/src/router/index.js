import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import ProductNewView from '@/views/ProductNewView.vue'
import ProductDetailView from '@/views/ProductDetailView.vue'
import PostLogsView from '@/views/PostLogsView.vue'
import SoldProductsView from '@/views/SoldProductsView.vue'
import SettingsView from '@/views/SettingsView.vue'
import ContentBankView from '@/views/ContentBankView.vue'
import ContentCaptureView from '@/views/ContentCaptureView.vue'
import ContentItemDetailView from '@/views/ContentItemDetailView.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomeView },
    { path: '/products/new', component: ProductNewView },
    { path: '/products/:id', component: ProductDetailView },
    { path: '/post-logs', component: PostLogsView },
    { path: '/sold', component: SoldProductsView },
    { path: '/settings', component: SettingsView },
    { path: '/content-bank', component: ContentBankView },
    { path: '/content-bank/capture', component: ContentCaptureView },
    { path: '/content-bank/:id', component: ContentItemDetailView },
  ],
})
