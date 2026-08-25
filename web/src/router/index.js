import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import ProductNewView from '@/views/ProductNewView.vue'
import ProductDetailView from '@/views/ProductDetailView.vue'
import PostLogsView from '@/views/PostLogsView.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomeView },
    { path: '/products/new', component: ProductNewView },
    { path: '/products/:id', component: ProductDetailView },
    { path: '/post-logs', component: PostLogsView },
  ],
})
