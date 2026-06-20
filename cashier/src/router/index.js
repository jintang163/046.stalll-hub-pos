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
  },
  {
    path: '/stall-cashier',
    name: 'StallCashier',
    component: () => import('@/views/stall-cashier/index.vue'),
    meta: { title: '摊位收银' }
  },
  {
    path: '/stall-report',
    name: 'StallReport',
    component: () => import('@/views/stall-report/index.vue'),
    meta: { title: '摊位报表' }
  },
  {
    path: '/queue-call',
    name: 'QueueCall',
    component: () => import('@/views/queue-call/index.vue'),
    meta: { title: '排队叫号' }
  },
  {
    path: '/stock-check',
    name: 'StockCheck',
    component: () => import('@/views/stock-check/index.vue'),
    meta: { title: '库存盘点' }
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
