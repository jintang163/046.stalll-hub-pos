import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Cashier',
    component: () => import('@/views/cashier/index.vue'),
    meta: { title: '收银台' }
  },
  {
    path: '/orders',
    name: 'Orders',
    component: () => import('@/views/orders/index.vue'),
    meta: { title: '订单管理' }
  },
  {
    path: '/sync',
    name: 'Sync',
    component: () => import('@/views/sync/index.vue'),
    meta: { title: '数据同步' }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/settings/index.vue'),
    meta: { title: '系统设置' }
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  document.title = to.meta.title ? `${to.meta.title} - 大排档收银系统` : '大排档收银系统'
  next()
})

export default router
